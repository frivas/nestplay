package youtube

import (
	"context"
	"testing"
)

func TestDisabledAsksForReconnect(t *testing.T) {
	page := Disabled().Status(context.Background())
	if page.State != Reconnect {
		t.Fatalf("state = %q", page.State)
	}
	if page.Message == "" {
		t.Fatal("reconnect needs a message")
	}
	if len(page.Items) != 0 {
		t.Fatal("disconnected YouTube must not invent recommendations")
	}
}

func TestVideoID(t *testing.T) {
	if !VideoID("abcdefghijk") {
		t.Fatal("11-character id")
	}
	if VideoID("short") || VideoID("not a video") || VideoID("../../../x") {
		t.Fatal("rejected bad ids")
	}
}

func TestNewMissingCookiesReconnects(t *testing.T) {
	page := New(t.TempDir() + "/missing-cookies.txt").Status(context.Background())
	if page.State != Reconnect || len(page.Items) != 0 {
		t.Fatalf("%+v", page)
	}
}
