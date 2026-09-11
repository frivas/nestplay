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

func TestConfigValidateRejectsDevModeAndPaid(t *testing.T) {
	ok := Config{TMDBPlan: PlanFree, OpenSubtitlesPlan: PlanFree}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	paid := Config{TMDBPlan: PlanFree, OpenSubtitlesPlan: Plan("vip")}
	if err := paid.Validate(); !errors.Is(err, ErrPaidPlan) {
		t.Fatalf("paid plan: %v", err)
	}
	dev := Config{TMDBPlan: PlanFree, OpenSubtitlesPlan: PlanFree, OpenSubtitlesDevMode: true}
	if err := dev.Validate(); !errors.Is(err, ErrDevMode) {
		t.Fatalf("dev mode: %v", err)
	}
}

func TestValidateFreeAccountRejectsVIPAndHighQuota(t *testing.T) {
	free := OpenSubtitlesAccount{AllowedDownloads: 5, RemainingDownloads: 5, RemainingKnown: true}
	if err := free.ValidateFreeAccount(); err != nil {
		t.Fatal(err)
	}
	if err := (OpenSubtitlesAccount{VIP: true, AllowedDownloads: 5}).ValidateFreeAccount(); !errors.Is(err, ErrPaidAccount) {
		t.Fatalf("vip: %v", err)
	}
	if err := (OpenSubtitlesAccount{Level: "VIP Member", AllowedDownloads: 5}).ValidateFreeAccount(); !errors.Is(err, ErrPaidAccount) {
		t.Fatalf("level: %v", err)
	}
	if err := (OpenSubtitlesAccount{AllowedDownloads: 1000}).ValidateFreeAccount(); !errors.Is(err, ErrPaidAccount) {
		t.Fatalf("1000/day: %v", err)
	}
}

func TestAuthorizeDownloadProtectsQuota(t *testing.T) {
	cfg := Config{TMDBPlan: PlanFree, OpenSubtitlesPlan: PlanFree}
	account := OpenSubtitlesAccount{AllowedDownloads: 5, RemainingDownloads: 5, RemainingKnown: true}
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
	spent := Budget{Day: now, Downloads: FreeOpenSubtitlesDownloadsPerDay}
	if err := AuthorizeDownload(cfg, account, spent, emptyLocal, hash, now); !errors.Is(err, ErrQuotaExhausted) {
		t.Fatalf("budget: %v", err)
	}
	reset := Budget{Day: now.Add(-24 * time.Hour), Downloads: FreeOpenSubtitlesDownloadsPerDay}
	if err := AuthorizeDownload(cfg, account, reset, emptyLocal, hash, now); err != nil {
		t.Fatalf("next UTC day should reset: %v", err)
	}
	tight := OpenSubtitlesAccount{AllowedDownloads: 5, RemainingDownloads: 0, RemainingKnown: true}
	if err := AuthorizeDownload(cfg, tight, budget, emptyLocal, hash, now); !errors.Is(err, ErrQuotaExhausted) {
		t.Fatalf("provider remaining: %v", err)
	}
}
