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
	"net/http"
	"os"
	"os/signal"
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
	return nestplayHandler(libraryRoot, youtube.Disabled())
}

func nestplayHandler(libraryRoot string, yt youtube.Adapter) http.Handler {
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
	mux.HandleFunc("GET /library", libraryHandler(libraryRoot))
	mux.HandleFunc("GET /library/{id}/artwork", artworkHandler(libraryRoot))
	mux.HandleFunc("GET /youtube/status", youtubeStatusHandler(yt))
	mux.HandleFunc("GET /youtube/home", youtubeHomeHandler(yt))
	mux.HandleFunc("GET /youtube/search", youtubeSearchHandler(yt))
	mux.HandleFunc("GET /youtube/{id}/artwork", youtubeArtworkHandler)
	mux.HandleFunc("POST /debug/log", debugLogHandler)
	return mux
}

type libraryItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Filename string `json:"filename"`
	Artwork  string `json:"artwork"`
}

type youtubeItem struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Artwork string `json:"artwork"`
}

func libraryHandler(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := catalog.List(root)
		if err != nil {
			http.Error(w, "library unavailable", http.StatusInternalServerError)
			return
		}
		out := make([]libraryItem, len(items))
		for i, item := range items {
			out[i] = libraryItem{
				ID: item.ID, Title: item.Title, Filename: item.Filename,
				Artwork: "/library/" + item.ID + "/artwork",
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(struct {
			Items []libraryItem `json:"items"`
		}{Items: out})
	}
}

func artworkHandler(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !catalogID(id) {
			http.NotFound(w, r)
			return
		}
		file, err := catalog.FileForID(root, id)
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

func youtubeStatusHandler(yt youtube.Adapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeYouTubePage(w, yt.Status(r.Context()))
	}
}

func youtubeHomeHandler(yt youtube.Adapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeYouTubePage(w, yt.Home(r.Context(), r.URL.Query().Get("cursor")))
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
	server := &http.Server{
		Addr: cfg.Listen, Handler: nestplayHandler(cfg.LibraryRoot, youtube.New(cfg.YouTubeCookies)),
		ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 45 * time.Second,
		IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	fmt.Printf("Nestplay health server listening on %s\nGET /health\nGET /library\nGET /library/{id}/artwork\nGET /youtube/status\nGET /youtube/home\nGET /youtube/search\nGET /youtube/{id}/artwork\nPOST /debug/log\n", cfg.Listen)
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
