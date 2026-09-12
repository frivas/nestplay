package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"nestplay.local/server/internal/artwork"
	"nestplay.local/server/internal/catalog"
	"nestplay.local/server/internal/youtube"
)

type config struct {
	Listen         string `json:"listen"`
	LibraryRoot    string `json:"library_root"`
	YouTubeCookies string `json:"youtube_cookies"`
}

const maxDebugBody = 64 << 10

func apiHandler(libraryRoot string) http.Handler {
	cat, err := catalog.OpenImmediate(libraryRoot, "")
	if err != nil {
		cat, _ = catalog.OpenImmediate("", "")
	}
	return nestplayHandler(cat, youtube.Disabled())
}

func nestplayHandler(cat *catalog.Catalog, yt youtube.Adapter) http.Handler {
	if cat == nil {
		cat, _ = catalog.OpenImmediate("", "")
	}
	if yt == nil {
		yt = youtube.Disabled()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(struct {
			Service    string `json:"service"`
			Status     string `json:"status"`
			APIVersion int    `json:"api_version"`
		}{"nestplay", "ok", 1})
	})
	mux.HandleFunc("GET /library", libraryHandler(cat))
	mux.HandleFunc("GET /library/events", libraryEventsHandler(cat))
	mux.HandleFunc("GET /library/{id}/artwork", artworkHandler(cat))
	mux.HandleFunc("GET /library/{id}/media", mediaHandler(cat))
	mux.HandleFunc("GET /youtube/status", youtubeStatusHandler(yt))
	mux.HandleFunc("GET /youtube/home", youtubeHomeHandler(yt))
	mux.HandleFunc("GET /youtube/search", youtubeSearchHandler(yt))
	mux.HandleFunc("GET /youtube/{id}/artwork", youtubeArtworkHandler)
	mux.HandleFunc("POST /debug/log", debugLogHandler)
	return mux
}

type libraryItem struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Filename   string `json:"filename"`
	Collection string `json:"collection"`
	Artwork    string `json:"artwork"`
}

type youtubeItem struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Artwork string `json:"artwork"`
}

func libraryHandler(cat *catalog.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items := cat.Items()
		out := make([]libraryItem, len(items))
		for i, item := range items {
			out[i] = libraryItem{
				ID: item.ID, Title: item.Title, Filename: item.Filename,
				Collection: item.Collection,
				Artwork:    "/library/" + item.ID + "/artwork",
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(struct {
			Items []libraryItem `json:"items"`
		}{Items: out})
	}
}

func libraryEventsHandler(cat *catalog.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "events unavailable", http.StatusInternalServerError)
			return
		}
		_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, ": connected\n\n"); err != nil {
			return
		}
		flusher.Flush()
		ch, unsub := cat.Subscribe()
		defer unsub()
		ping := time.NewTicker(15 * time.Second)
		defer ping.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-ping.C:
				if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
					return
				}
				flusher.Flush()
			case _, ok := <-ch:
				if !ok {
					return
				}
				if _, err := io.WriteString(w, "data: {\"changed\":true}\n\n"); err != nil {
					return
				}
				flusher.Flush()
			}
		}
	}
}

func artworkHandler(cat *catalog.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !catalogID(id) {
			http.NotFound(w, r)
			return
		}
		file, err := cat.FileByID(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		data, err := artwork.JPEG(id, file.Path)
		if err != nil || !bytes.HasPrefix(data, []byte{0xFF, 0xD8}) {
			http.Error(w, "artwork unavailable", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "private, max-age=3600")
		http.ServeContent(w, r, id+".jpg", time.Time{}, bytes.NewReader(data))
	}
}

func mediaHandler(cat *catalog.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !catalogID(id) {
			http.NotFound(w, r)
			return
		}
		file, err := cat.FileByID(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		f, err := os.Open(file.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil {
			http.Error(w, "media unavailable", http.StatusInternalServerError)
			return
		}
		_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
		if ctype := mime.TypeByExtension(strings.ToLower(filepath.Ext(file.Item.Filename))); ctype != "" {
			w.Header().Set("Content-Type", ctype)
		}
		w.Header().Set("Cache-Control", "private")
		http.ServeContent(w, r, file.Item.Filename, stat.ModTime(), f)
	}
}

func youtubeStatusHandler(yt youtube.Adapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeYouTubePage(w, yt.Status(r.Context()))
	}
}

func youtubeHomeHandler(yt youtube.Adapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeYouTubePage(w, yt.Home(r.Context(), r.URL.Query().Get("cursor"), r.URL.Query().Get("refresh") == "1"))
	}
}

func youtubeSearchHandler(yt youtube.Adapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeYouTubePage(w, yt.Search(r.Context(), r.URL.Query().Get("q"), r.URL.Query().Get("cursor")))
	}
}

func writeYouTubePage(w http.ResponseWriter, page youtube.Page) {
	items := make([]youtubeItem, 0, len(page.Items))
	for _, item := range page.Items {
		if !youtube.VideoID(item.ID) {
			continue
		}
		items = append(items, youtubeItem{
			ID: item.ID, Title: item.Title,
			Artwork: "/youtube/" + item.ID + "/artwork",
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(struct {
		State   string        `json:"state"`
		Message string        `json:"message,omitempty"`
		Items   []youtubeItem `json:"items"`
		Cursor  string        `json:"cursor,omitempty"`
	}{string(page.State), page.Message, items, page.Cursor})
}

func youtubeArtworkHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !youtube.VideoID(id) {
		http.NotFound(w, r)
		return
	}
	data := youtubeThumbnail(id)
	if !bytes.HasPrefix(data, []byte{0xFF, 0xD8}) {
		data = artwork.Placeholder()
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private, max-age=3600")
	http.ServeContent(w, r, id+".jpg", time.Time{}, bytes.NewReader(data))
}

func youtubeThumbnail(id string) []byte {
	client := &http.Client{Timeout: 6 * time.Second}
	for _, name := range []string{"hq720.jpg", "maxresdefault.jpg", "hqdefault.jpg"} {
		addr := "https://i.ytimg.com/vi/" + id + "/" + name
		req, err := http.NewRequest(http.MethodGet, addr, nil) // #nosec G107,G704 -- id is a validated 11-character YouTube video id
		if err != nil {
			continue
		}
		resp, err := client.Do(req) // #nosec G704 -- host is i.ytimg.com; id is a validated YouTube video id
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusOK && bytes.HasPrefix(body, []byte{0xFF, 0xD8}) {
			return body
		}
	}
	return artwork.Placeholder()
}

func catalogID(id string) bool {
	if len(id) != 32 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if c >= '0' && c <= '9' || c >= 'a' && c <= 'f' {
			continue
		}
		return false
	}
	return true
}

func sanitizeLog(s string) string {
	return strings.NewReplacer("\n", "", "\r", "", "\x00", "").Replace(s)
}

func debugLogHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxDebugBody)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "debug log too large", http.StatusRequestEntityTooLarge)
		return
	}
	var event struct {
		Source  string `json:"source"`
		Level   string `json:"level"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &event); err != nil || strings.TrimSpace(event.Message) == "" {
		http.Error(w, "invalid debug log", http.StatusBadRequest)
		return
	}
	if event.Source == "" {
		event.Source = "tvos"
	}
	if event.Level == "" {
		event.Level = "debug"
	}
	log.Printf(
		"%s %s %s",
		sanitizeLog(event.Source),
		sanitizeLog(event.Level),
		sanitizeLog(event.Message),
	)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	configPath := flag.String("config", "", "optional JSON configuration file")
	listen := flag.String("listen", "", "listen address; overrides configuration")
	libraryRoot := flag.String("library-root", "", "watched folder; overrides configuration")
	youtubeCookies := flag.String("youtube-cookies", "", "YouTube cookie file; overrides configuration")
	flag.Parse()
	cfg := config{Listen: "0.0.0.0:8765"}
	if *configPath != "" {
		data, err := os.ReadFile(*configPath)
		if err != nil {
			log.Fatal(err)
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Fatal(err)
		}
	}
	if *listen != "" {
		cfg.Listen = *listen
	}
	if *libraryRoot != "" {
		cfg.LibraryRoot = *libraryRoot
	}
	if *youtubeCookies != "" {
		cfg.YouTubeCookies = *youtubeCookies
	}
	if cfg.YouTubeCookies == "" {
		cfg.YouTubeCookies = youtube.DefaultCookieFile()
	}
	if cfg.Listen == "" {
		log.Fatal("listen must not be empty")
	}
	cat, err := catalog.Open(cfg.LibraryRoot, catalog.DefaultDBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer cat.Close()
	server := &http.Server{
		Addr: cfg.Listen, Handler: nestplayHandler(cat, youtube.Cached(youtube.New(cfg.YouTubeCookies), youtube.DefaultHomeCacheFile(), cfg.YouTubeCookies)),
		ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 45 * time.Second,
		IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go cat.Watch(ctx)
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	fmt.Printf("Nestplay health server listening on %s\nGET /health\nGET /library\nGET /library/events\nGET /library/{id}/artwork\nGET /library/{id}/media\nGET /youtube/status\nGET /youtube/home\nGET /youtube/search\nGET /youtube/{id}/artwork\nPOST /debug/log\n", cfg.Listen)
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		log.Fatal(err)
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			log.Print(err)
		}
	}
}
