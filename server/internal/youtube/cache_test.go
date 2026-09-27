package youtube

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type countingHome struct {
	Fake
	n int
}

func (c *countingHome) Home(ctx context.Context, cursor string, refresh bool) Page {
	c.n++
	return c.Fake.Home(ctx, cursor, refresh)
}

func TestHomeCacheSkipsInnerTubeUntilRefresh(t *testing.T) {
	dir := t.TempDir()
	cookie := filepath.Join(dir, "cookies.txt")
	if err := os.WriteFile(cookie, []byte("SID=abc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	inner := &countingHome{Fake: Fake{Pages: map[string]Page{
		"home:": {State: Connected, Items: []Item{{ID: "abcdefghijk", Title: "Heat"}}, Cursor: "next"},
	}}}
	home := Cached(inner, filepath.Join(dir, "home.json"), cookie)
	ctx := context.Background()
	first := home.Home(ctx, "", false)
	if first.State != Connected || len(first.Items) != 0 || first.Message != refreshHomeMessage {
		t.Fatalf("empty cache %+v", first)
	}
	if inner.n != 0 {
		t.Fatal("opening Home must not call YouTube")
	}
	refreshed := home.Home(ctx, "", true)
	if refreshed.State != Connected || len(refreshed.Items) != 1 || inner.n != 1 {
		t.Fatalf("refresh %+v calls=%d", refreshed, inner.n)
	}
	again := home.Home(ctx, "", false)
	if again.State != Connected || len(again.Items) != 1 || again.Cursor != "next" || inner.n != 1 {
		t.Fatalf("cached %+v calls=%d", again, inner.n)
	}
}

func TestHomeCacheKeepsShelfWhenRefreshFails(t *testing.T) {
	dir := t.TempDir()
	cookie := filepath.Join(dir, "cookies.txt")
	if err := os.WriteFile(cookie, []byte("SID=abc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	inner := &countingHome{Fake: Fake{Pages: map[string]Page{
		"home:": {State: Connected, Items: []Item{{ID: "abcdefghijk", Title: "Heat"}}},
	}}}
	home := Cached(inner, filepath.Join(dir, "home.json"), cookie)
	ctx := context.Background()
	_ = home.Home(ctx, "", true)
	inner.Pages["home:"] = Page{State: Unavailable, Message: "YouTube Home could not be loaded."}
	failed := home.Home(ctx, "", true)
	if failed.State != Connected || len(failed.Items) != 1 {
		t.Fatalf("keep shelf %+v", failed)
	}
}

func TestHomeCacheReconnectsWithoutCookies(t *testing.T) {
	dir := t.TempDir()
	home := Cached(Disabled(), filepath.Join(dir, "home.json"), filepath.Join(dir, "missing-cookies.txt"))
	page := home.Home(context.Background(), "", false)
	if page.State != Reconnect || len(page.Items) != 0 {
		t.Fatalf("%+v", page)
	}
}
