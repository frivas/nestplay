package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthContract(t *testing.T) {
	w := httptest.NewRecorder()
	apiHandler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
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
		{http.MethodGet, "/debug/log", http.StatusMethodNotAllowed},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			apiHandler().ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("got %d want %d", w.Code, tc.status)
			}
		})
	}
}

func TestDebugLogAcceptsJSON(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/debug/log", strings.NewReader(
		`{"source":"tvos","level":"info","message":"check started"}`))
	apiHandler().ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestDebugLogRejectsInvalidAndOversize(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		apiHandler().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/debug/log", strings.NewReader("not-json")))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got %d want 400", w.Code)
		}
	})
	t.Run("missing message", func(t *testing.T) {
		w := httptest.NewRecorder()
		apiHandler().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/debug/log", strings.NewReader(`{"level":"info"}`)))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got %d want 400", w.Code)
		}
	})
	t.Run("too large", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"message":"` + strings.Repeat("x", maxDebugBody) + `"}`
		apiHandler().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/debug/log", strings.NewReader(body)))
		if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
			t.Fatalf("got %d want 413 or 400", w.Code)
		}
	})
}
