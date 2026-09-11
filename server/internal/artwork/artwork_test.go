package artwork

import "testing"

func TestPlaceholderIsJPEG(t *testing.T) {
	data := placeholder()
	if !jpegSOI(data) {
		t.Fatal("placeholder is not JPEG")
	}
	if len(data) < 800 {
		t.Fatalf("placeholder too small: %d", len(data))
	}
}
