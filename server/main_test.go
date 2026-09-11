package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHealthContract(t *testing.T) {
	w := httptest.NewRecorder()
	apiHandler("").ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
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
		{http.MethodGet, "/library/extra", http.StatusNotFound},
		{http.MethodPost, "/library", http.StatusMethodNotAllowed},
		{http.MethodPost, "/health", http.StatusMethodNotAllowed},
		{http.MethodGet, "/debug/log", http.StatusMethodNotAllowed},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			apiHandler("").ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("got %d want %d", w.Code, tc.status)
			}
		})
	}
}

func TestSanitizeLogStripsLineBreaks(t *testing.T) {
	got := sanitizeLog("ok\r\ninjected\x00")
	if got != "okinjected" {
		t.Fatalf("got %q", got)
	}
}

func TestDebugLogAcceptsJSON(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/debug/log", strings.NewReader(
		`{"source":"tvos","level":"info","message":"check started"}`))
	apiHandler("").ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestDebugLogRejectsInvalidAndOversize(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		apiHandler("").ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/debug/log", strings.NewReader("not-json")))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got %d want 400", w.Code)
		}
	})
	t.Run("missing message", func(t *testing.T) {
		w := httptest.NewRecorder()
		apiHandler("").ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/debug/log", strings.NewReader(`{"level":"info"}`)))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got %d want 400", w.Code)
		}
	})
	t.Run("too large", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"message":"` + strings.Repeat("x", maxDebugBody) + `"}`
		apiHandler("").ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/debug/log", strings.NewReader(body)))
		if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
			t.Fatalf("got %d want 413 or 400", w.Code)
		}
	})
}

func TestLibraryJSON(t *testing.T) {
	t.Run("empty root", func(t *testing.T) {
		w := httptest.NewRecorder()
		apiHandler("").ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/library", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("library must not be cached")
		}
		var body struct {
			Items []struct {
				ID       string `json:"id"`
				Title    string `json:"title"`
				Filename string `json:"filename"`
			} `json:"items"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Items) != 0 {
			t.Fatalf("got %+v", body.Items)
		}
	})
	t.Run("watched folder", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "Heat.mkv"), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		apiHandler(root).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/library", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		var body struct {
			Items []struct {
				ID       string `json:"id"`
				Title    string `json:"title"`
				Filename string `json:"filename"`
				Artwork  string `json:"artwork"`
			} `json:"items"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Items) != 1 || body.Items[0].Title != "Heat" || body.Items[0].Filename != "Heat.mkv" {
			t.Fatalf("got %+v", body.Items)
		}
		wantArt := "/library/" + body.Items[0].ID + "/artwork"
		if body.Items[0].Artwork != wantArt {
			t.Fatalf("artwork = %q", body.Items[0].Artwork)
		}
		if strings.Contains(w.Body.String(), root) {
			t.Fatal("watched-folder path leaked on the wire")
		}
		art := httptest.NewRecorder()
		apiHandler(root).ServeHTTP(art, httptest.NewRequest(http.MethodGet, wantArt, nil))
		if art.Code != http.StatusOK {
			t.Fatalf("artwork status = %d", art.Code)
		}
		if art.Header().Get("Content-Type") != "image/jpeg" {
			t.Fatal("artwork must be JPEG")
		}
		if len(art.Body.Bytes()) < 2 || art.Body.Bytes()[0] != 0xFF || art.Body.Bytes()[1] != 0xD8 {
			t.Fatal("artwork is not a JPEG")
		}
		missing := httptest.NewRecorder()
		apiHandler(root).ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/library/0123456789abcdef0123456789abcdef/artwork", nil))
		if missing.Code != http.StatusNotFound {
			t.Fatalf("unknown artwork status = %d", missing.Code)
		}
	})
}
