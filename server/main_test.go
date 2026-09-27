package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nestplay.local/server/internal/catalog"
	"nestplay.local/server/internal/downloads"
	"nestplay.local/server/internal/youtube"
)

func mustCatalog(t *testing.T, root string) *catalog.Catalog {
	t.Helper()
	cat, err := catalog.OpenImmediate(root, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cat.Close() })
	return cat
}

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
		{http.MethodPost, "/library/0123456789abcdef0123456789abcdef/media", http.StatusMethodNotAllowed},
		{http.MethodPost, "/health", http.StatusMethodNotAllowed},
		{http.MethodPost, "/youtube/home", http.StatusMethodNotAllowed},
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
				ID         string `json:"id"`
				Title      string `json:"title"`
				Filename   string `json:"filename"`
				Collection string `json:"collection"`
				Artwork    string `json:"artwork"`
			} `json:"items"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Items) != 1 || body.Items[0].Title != "Heat" || body.Items[0].Filename != "Heat.mkv" || body.Items[0].Collection != "" {
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
	t.Run("collection is not a mac path", func(t *testing.T) {
		root := t.TempDir()
		dir := filepath.Join(root, "Mateo", "PJ Masks")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "09 - Gatuno.mp4"), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		apiHandler(root).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/library", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if strings.Contains(w.Body.String(), root) {
			t.Fatal("watched-folder path leaked on the wire")
		}
		var body struct {
			Items []struct {
				Collection string `json:"collection"`
				Filename   string `json:"filename"`
			} `json:"items"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Items) != 1 || body.Items[0].Filename != "09 - Gatuno.mp4" || body.Items[0].Collection != "Mateo/PJ Masks" {
			t.Fatalf("got %+v", body.Items)
		}
	})
}

func TestLibraryMediaRange(t *testing.T) {
	root := t.TempDir()
	payload := bytes.Repeat([]byte("mp4-"), 256*1024)
	if err := os.WriteFile(filepath.Join(root, "Heat.mp4"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	cat := mustCatalog(t, root)
	items := cat.Items()
	if len(items) != 1 {
		t.Fatalf("catalog: %+v", items)
	}
	h := nestplayHandler(cat, youtube.Disabled(), nil)
	srv := httptest.NewUnstartedServer(h)
	srv.Config.WriteTimeout = 80 * time.Millisecond
	srv.Start()
	t.Cleanup(srv.Close)

	media := srv.URL + "/library/" + items[0].ID + "/media"
	req, err := http.NewRequest(http.MethodGet, media, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Range", "bytes=0-15")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusPartialContent {
		t.Fatalf("range status = %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "video/mp4" {
		t.Fatalf("type = %q", resp.Header.Get("Content-Type"))
	}
	if !bytes.Equal(body, payload[:16]) {
		t.Fatalf("range body %q", body)
	}
	if strings.Contains(resp.Header.Get("Content-Disposition"), root) || strings.Contains(string(body), root) {
		t.Fatal("watched-folder path leaked on the wire")
	}

	full, err := http.Get(media)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(full.Body)
	_ = full.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if full.StatusCode != http.StatusOK || !bytes.Equal(got, payload) {
		t.Fatalf("full get status=%d len=%d", full.StatusCode, len(got))
	}

	missing := httptest.NewRecorder()
	h.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/library/0123456789abcdef0123456789abcdef/media", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("unknown media status = %d", missing.Code)
	}
}

func TestLibraryEventsSignalsChange(t *testing.T) {
	root := t.TempDir()
	cat := mustCatalog(t, root)
	srv := httptest.NewServer(nestplayHandler(cat, youtube.Disabled(), nil))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/library/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		t.Fatalf("type = %q", resp.Header.Get("Content-Type"))
	}

	got := make(chan string, 1)
	ready := make(chan struct{})
	go func() {
		buf := make([]byte, 0, 512)
		tmp := make([]byte, 256)
		connected := false
		for {
			n, readErr := resp.Body.Read(tmp)
			buf = append(buf, tmp[:n]...)
			body := string(buf)
			if !connected && strings.Contains(body, ": connected") {
				connected = true
				close(ready)
			}
			if strings.Contains(body, `data: {"changed":true}`) {
				got <- body
				return
			}
			if readErr != nil {
				if !connected {
					close(ready)
				}
				got <- "err:" + readErr.Error() + " body:" + body
				return
			}
		}
	}()
	select {
	case <-ready:
	case <-ctx.Done():
		t.Fatal("timed out waiting for events stream")
	}
	if err := os.WriteFile(filepath.Join(root, "Heat.mkv"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := cat.Scan(); err != nil {
		t.Fatal(err)
	}
	select {
	case body := <-got:
		if !strings.Contains(body, `data: {"changed":true}`) {
			t.Fatalf("events: %s", body)
		}
		if strings.Contains(body, root) {
			t.Fatal("watched-folder path leaked on the wire")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for catalog change")
	}
}

func TestYouTubeJSON(t *testing.T) {
	t.Run("reconnect without cookies", func(t *testing.T) {
		w := httptest.NewRecorder()
		apiHandler("").ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/youtube/status", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("youtube must not be cached")
		}
		var body struct {
			State   string `json:"state"`
			Message string `json:"message"`
			Items   []any  `json:"items"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.State != "reconnect_required" || len(body.Items) != 0 || body.Message == "" {
			t.Fatalf("got %+v", body)
		}
	})
	t.Run("home continuation and search", func(t *testing.T) {
		fake := &youtube.Fake{
			Pages: map[string]youtube.Page{
				"status": {State: youtube.Connected},
				"home:": {
					State:  youtube.Connected,
					Items:  []youtube.Item{{ID: "abcdefghijk", Title: "Heat"}},
					Cursor: "next",
				},
				"home:next": {
					State: youtube.Connected,
					Items: []youtube.Item{{ID: "lmnopqrstuv", Title: "More"}},
				},
				"search:heat:": {
					State: youtube.Connected,
					Items: []youtube.Item{{ID: "abcdefghijk", Title: "Heat"}},
				},
			},
		}
		h := nestplayHandler(mustCatalog(t, ""), fake, nil)
		home := httptest.NewRecorder()
		h.ServeHTTP(home, httptest.NewRequest(http.MethodGet, "/youtube/home", nil))
		var first struct {
			State  string        `json:"state"`
			Items  []youtubeItem `json:"items"`
			Cursor string        `json:"cursor"`
		}
		if err := json.Unmarshal(home.Body.Bytes(), &first); err != nil {
			t.Fatal(err)
		}
		if first.State != "connected" || len(first.Items) != 1 || first.Items[0].Artwork != "/youtube/abcdefghijk/artwork" || first.Cursor != "next" {
			t.Fatalf("home %+v", first)
		}
		more := httptest.NewRecorder()
		h.ServeHTTP(more, httptest.NewRequest(http.MethodGet, "/youtube/home?cursor=next", nil))
		if !strings.Contains(more.Body.String(), "lmnopqrstuv") {
			t.Fatalf("continuation %s", more.Body.String())
		}
		search := httptest.NewRecorder()
		h.ServeHTTP(search, httptest.NewRequest(http.MethodGet, "/youtube/search?q=heat", nil))
		if !strings.Contains(search.Body.String(), "abcdefghijk") {
			t.Fatalf("search %s", search.Body.String())
		}
		missing := httptest.NewRecorder()
		h.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/youtube/nope/artwork", nil))
		if missing.Code != http.StatusNotFound {
			t.Fatalf("artwork status = %d", missing.Code)
		}
	})
}

func TestQueueEnqueueListsWithoutPaths(t *testing.T) {
	root := t.TempDir()
	q := downloads.Open(root, t.TempDir(), nil, nil, nil)
	h := nestplayHandler(mustCatalog(t, root), youtube.Disabled(), q)
	post := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/queue", strings.NewReader(
		`{"items":[{"id":"dQw4w9WgXcQ","title":"Heat"}]}`))
	h.ServeHTTP(post, req)
	if post.Code != http.StatusOK {
		t.Fatalf("status = %d", post.Code)
	}
	if strings.Contains(post.Body.String(), root) {
		t.Fatal("path leaked")
	}
	var body struct {
		Items []struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"items"`
	}
	if err := json.Unmarshal(post.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != "dQw4w9WgXcQ" || body.Items[0].State != "queued" {
		t.Fatalf("%+v", body.Items)
	}
	get := httptest.NewRecorder()
	h.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/queue", nil))
	if get.Code != http.StatusOK || !strings.Contains(get.Body.String(), "dQw4w9WgXcQ") {
		t.Fatalf("get %s", get.Body.String())
	}
}
