package youtube

import (
	"strings"
	"testing"
)

func TestCookieHeaderKeepsYouTubeNetscapePairs(t *testing.T) {
	raw := "# Netscape HTTP Cookie File\n" +
		".youtube.com\tTRUE\t/\tTRUE\t0\tSID\tabc\n" +
		".other.example\tTRUE\t/\tTRUE\t0\tsecret\tnosey\n" +
		".google.com\tTRUE\t/\tTRUE\t0\tHSID\tdef\n"
	got := cookieHeader(raw)
	if got != "SID=abc; HSID=def" {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(got, "nosey") || strings.Contains(got, "secret") {
		t.Fatal("unrelated site cookies must not be imported")
	}
}

func TestCookieHeaderPassesThroughHeaderForm(t *testing.T) {
	got := cookieHeader("SID=abc; HSID=def\n")
	if got != "SID=abc; HSID=def" {
		t.Fatalf("got %q", got)
	}
}
