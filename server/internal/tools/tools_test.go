package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveUsesDirectoryJoin(t *testing.T) {
	dir := t.TempDir()
	ffmpeg := filepath.Join(dir, "ffmpeg")
	if err := os.WriteFile(ffmpeg, []byte("x"), 0o700); err != nil {
		t.Fatal(err)
	}
	Configure("", dir)
	t.Cleanup(func() { Configure("", "") })
	got, err := FFmpeg()
	if err != nil || got != ffmpeg {
		t.Fatalf("got %q %v", got, err)
	}
}

func TestResolveUsesExactYTDLPFile(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "yt-dlp")
	if err := os.WriteFile(bin, []byte("x"), 0o700); err != nil {
		t.Fatal(err)
	}
	Configure(bin, "")
	t.Cleanup(func() { Configure("", "") })
	got, err := YTDLP()
	if err != nil || got != bin {
		t.Fatalf("got %q %v", got, err)
	}
}
