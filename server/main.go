package main

import (
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

	"nestplay.local/server/internal/providers"
)

type config struct {
	Listen string `json:"listen"`
}

const maxDebugBody = 64 << 10

func apiHandler() http.Handler {
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
	mux.HandleFunc("POST /debug/log", debugLogHandler)
	return mux
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
	if cfg.Listen == "" {
		log.Fatal("listen must not be empty")
	}
	if _, secretsPath, err := providers.LoadOptionalConfig(); err != nil {
		log.Fatal(err)
	} else if secretsPath != "" {
		log.Printf("using unpaid provider policy from %s", secretsPath)
	}
	server := &http.Server{
		Addr: cfg.Listen, Handler: apiHandler(),
		ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	fmt.Printf("Nestplay health server listening on %s\nGET /health\nPOST /debug/log\n", cfg.Listen)
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
