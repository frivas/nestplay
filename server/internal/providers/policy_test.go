package providers

import (
	"errors"
	"testing"
	"time"
)

func TestParsePlanRejectsPaidAndUnknown(t *testing.T) {
	if p, err := ParsePlan(""); err != nil || p != PlanFree {
		t.Fatalf("empty: %q %v", p, err)
	}
	if p, err := ParsePlan("Developer"); err != nil || p != PlanFree {
		t.Fatalf("developer: %q %v", p, err)
	}
	for _, paid := range []string{"vip", "Light", "startup", "BASIC", "premium", "pro", "enterprise", "commercial"} {
		if _, err := ParsePlan(paid); !errors.Is(err, ErrPaidPlan) {
			t.Fatalf("%s: want ErrPaidPlan, got %v", paid, err)
		}
	}
	if _, err := ParsePlan("gold"); !errors.Is(err, ErrPaidPlan) {
		t.Fatalf("unknown: %v", err)
	}
}

func TestConfigValidateAllowsUnderDevRejectsPaid(t *testing.T) {
	ok := defaultConfig()
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	if ok.DownloadCeiling() != FreeOpenSubtitlesDevDownloadsPerDay {
		t.Fatalf("under-dev ceiling = %d", ok.DownloadCeiling())
	}
	if !ok.AllowsAnonymousDownload() {
		t.Fatal("under-dev consumer downloads without JWT")
	}
	paid := Config{TMDBPlan: PlanFree, OpenSubtitlesPlan: Plan("vip")}
	if err := paid.Validate(); !errors.Is(err, ErrPaidPlan) {
		t.Fatalf("paid plan: %v", err)
	}
	anonOnly := Config{TMDBPlan: PlanFree, OpenSubtitlesPlan: PlanFree, OpenSubtitlesAnonymous: true}
	if anonOnly.DownloadCeiling() != FreeOpenSubtitlesAnonymousDownloadsPerDay {
		t.Fatalf("anonymous ceiling = %d", anonOnly.DownloadCeiling())
	}
}

func TestCheckAPIHostRejectsVIP(t *testing.T) {
	if err := CheckAPIHost("api.opensubtitles.com"); err != nil {
		t.Fatal(err)
	}
	if err := CheckAPIHost("https://vip-api.opensubtitles.com"); !errors.Is(err, ErrVIPEndpoint) {
		t.Fatalf("vip host: %v", err)
	}
}

func TestEncodeQuerySortsAndUsesPlus(t *testing.T) {
	got := EncodeQuery(map[string]string{"query": "The Matrix", "languages": "es"})
	if got != "languages=es&query=The+Matrix" {
		t.Fatalf("got %q", got)
	}
}

func TestValidateFreeAccountRejectsVIPAndHighQuota(t *testing.T) {
	cfg := defaultConfig()
	free := OpenSubtitlesAccount{AllowedDownloads: 100, RemainingDownloads: 100, RemainingKnown: true}
	if err := free.ValidateFreeAccount(cfg); err != nil {
		t.Fatal(err)
	}
	if err := (OpenSubtitlesAccount{VIP: true, AllowedDownloads: 100}).ValidateFreeAccount(cfg); !errors.Is(err, ErrPaidAccount) {
		t.Fatalf("vip: %v", err)
	}
	if err := (OpenSubtitlesAccount{Level: "VIP Member", AllowedDownloads: 100}).ValidateFreeAccount(cfg); !errors.Is(err, ErrPaidAccount) {
		t.Fatalf("level: %v", err)
	}
	if err := (OpenSubtitlesAccount{AllowedDownloads: 1000}).ValidateFreeAccount(cfg); !errors.Is(err, ErrPaidAccount) {
		t.Fatalf("1000/day: %v", err)
	}
}

func TestAuthorizeDownloadProtectsQuota(t *testing.T) {
	cfg := defaultConfig()
	account := OpenSubtitlesAccount{AllowedDownloads: 100, RemainingDownloads: 100, RemainingKnown: true}
	now := time.Date(2026, 9, 11, 16, 0, 0, 0, time.UTC)
	hash := SubtitleCandidate{HashMatch: true}
	emptyLocal := LocalSpanish{}
	budget := Budget{Day: now, Downloads: 0}

	if err := AuthorizeDownload(cfg, account, budget, emptyLocal, hash, now); err != nil {
		t.Fatalf("hash match should spend quota: %v", err)
	}
	if err := AuthorizeDownload(cfg, account, budget, LocalSpanish{EmbeddedFull: true}, hash, now); !errors.Is(err, ErrLocalSpanish) {
		t.Fatalf("embedded: %v", err)
	}
	if err := AuthorizeDownload(cfg, account, budget, LocalSpanish{SidecarFull: true}, hash, now); !errors.Is(err, ErrLocalSpanish) {
		t.Fatalf("sidecar: %v", err)
	}
	if err := AuthorizeDownload(cfg, account, budget, emptyLocal, SubtitleCandidate{MachineTranslated: true, HashMatch: true}, now); !errors.Is(err, ErrMachineTranslated) {
		t.Fatalf("mt: %v", err)
	}
	if err := AuthorizeDownload(cfg, account, budget, emptyLocal, SubtitleCandidate{}, now); !errors.Is(err, ErrWeakMatch) {
		t.Fatalf("name-only: %v", err)
	}
	spent := Budget{Day: now, Downloads: cfg.DownloadCeiling()}
	if err := AuthorizeDownload(cfg, account, spent, emptyLocal, hash, now); !errors.Is(err, ErrQuotaExhausted) {
		t.Fatalf("budget: %v", err)
	}
	reset := Budget{Day: now.Add(-24 * time.Hour), Downloads: cfg.DownloadCeiling()}
	if err := AuthorizeDownload(cfg, account, reset, emptyLocal, hash, now); err != nil {
		t.Fatalf("next UTC day should reset: %v", err)
	}
	tight := OpenSubtitlesAccount{AllowedDownloads: 100, RemainingDownloads: 0, RemainingKnown: true}
	if err := AuthorizeDownload(cfg, tight, budget, emptyLocal, hash, now); !errors.Is(err, ErrQuotaExhausted) {
		t.Fatalf("provider remaining: %v", err)
	}
	mid := Budget{Day: now, Downloads: 5}
	if err := AuthorizeDownload(cfg, account, mid, emptyLocal, hash, now); err != nil {
		t.Fatalf("under-dev still has quota after 5: %v", err)
	}
}
