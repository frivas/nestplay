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
)

type config struct {
	Listen      string `json:"listen"`
	LibraryRoot string `json:"library_root"`
}

const maxDebugBody = 64 << 10

func apiHandler(libraryRoot string) http.Handler {
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
	mux.HandleFunc("POST /debug/log", debugLogHandler)
	return mux
}

type libraryItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Filename string `json:"filename"`
	Artwork  string `json:"artwork"`
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
	if cfg.Listen == "" {
		log.Fatal("listen must not be empty")
	}
	server := &http.Server{
		Addr: cfg.Listen, Handler: apiHandler(cfg.LibraryRoot),
		ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	fmt.Printf("Nestplay health server listening on %s\nGET /health\nGET /library\nGET /library/{id}/artwork\nPOST /debug/log\n", cfg.Listen)
	select {
	case err := <-done:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			log.Print(err)
		}
	}
}
