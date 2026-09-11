package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthContract(t *testing.T) {
	w := httptest.NewRecorder()
	healthHandler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Fatal("missing JSON type")
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("health must not be cached")
	}
	var body struct {
		Service string `json:"service"`
		Status  string `json:"status"`
		Version int    `json:"api_version"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Service != "nestplay" || body.Status != "ok" || body.Version != 1 {
		t.Fatalf("wrong contract: %+v", body)
	}
}

func TestUnsupportedRoutes(t *testing.T) {
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{http.MethodGet, "/", http.StatusNotFound},
		{http.MethodGet, "/health/extra", http.StatusNotFound},
		{http.MethodPost, "/health", http.StatusMethodNotAllowed},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			healthHandler().ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("got %d want %d", w.Code, tc.status)
			}
		})
	}
}
