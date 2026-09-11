package artwork

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPlaceholderIsJPEG(t *testing.T) {
	data := placeholder()
	if !jpegSOI(data) {
		t.Fatal("placeholder is not JPEG")
	}
	if len(data) < 800 {
		t.Fatalf("placeholder too small: %d", len(data))
	}
}

func TestSeekSecondsLeavesSharedIntro(t *testing.T) {
	cases := []struct {
		duration float64
		min      float64
		max      float64
	}{
		{0, 45, 45},
		{8, 2, 4},
		{49, 15, 20},
		{50, 45, 45},
		{300, 54, 54},
		{22 * 60, 180, 180},
	}
	for _, c := range cases {
		got := seekSeconds(c.duration)
		if got == 8 && c.duration != 8 {
			t.Fatalf("duration %v seeked to the shared 8s intro", c.duration)
		}
		if got < c.min || got > c.max {
			t.Fatalf("duration %v: seek = %v, want %v–%v", c.duration, got, c.min, c.max)
		}
	}
}

func TestCacheFileBustsIntroStills(t *testing.T) {
	t.Setenv("NESTPLAY_ARTWORK_CACHE", t.TempDir())
	path, err := cacheFile("abc", time.Unix(1_700_000_000, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.Base(path), "-v3.jpg") {
		t.Fatalf("cache name %q must bust 8s stills", filepath.Base(path))
	}
}
