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

func TestCookieHeaderPrefersYouTubeHostForDuplicateNames(t *testing.T) {
	raw := ".google.com\tTRUE\t/\tTRUE\t0\tSAPISID\tgoogle\n" +
		".youtube.com\tTRUE\t/\tTRUE\t0\tSAPISID\tyoutube\n"
	got := cookieHeader(raw)
	if got != "SAPISID=youtube" {
		t.Fatalf("got %q", got)
	}
	raw = ".youtube.com\tTRUE\t/\tTRUE\t0\tSAPISID\tyoutube\n" +
		".google.com\tTRUE\t/\tTRUE\t0\tSAPISID\tgoogle\n"
	got = cookieHeader(raw)
	if got != "SAPISID=youtube" {
		t.Fatalf("got %q", got)
	}
}

func TestCookieHeaderPassesThroughHeaderForm(t *testing.T) {
	got := cookieHeader("SID=abc; HSID=def\n")
	if got != "SID=abc; HSID=def" {
		t.Fatalf("got %q", got)
	}
}
