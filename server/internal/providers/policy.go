package providers

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Unpaid OpenSubtitles.com consumer row from their Consumers pricing table
// (Free: $0, 5 requests/s, 5 downloads/24h). Nestplay never opts into Light
// through Enterprise, VIP, or the 100-download/day dev_mode test bypass.
const (
	FreeOpenSubtitlesDownloadsPerDay = 5
	FreeOpenSubtitlesRequestsPerSec  = 5
	freeOpenSubtitlesAccountCeiling  = 20
)

var (
	ErrPaidPlan          = errors.New("nestplay PoC allows only unpaid provider plans")
	ErrPaidAccount       = errors.New("refusing a paid, VIP, or Pro OpenSubtitles account")
	ErrDevMode           = errors.New("OpenSubtitles dev_mode is disabled; it inflates free-tier quota")
	ErrQuotaExhausted    = errors.New("free OpenSubtitles download quota is exhausted")
	ErrLocalSpanish      = errors.New("local Spanish subtitles already exist; skipping download")
	ErrMachineTranslated = errors.New("refusing machine-translated subtitles")
	ErrWeakMatch         = errors.New("refusing a weak subtitle match to preserve free quota")
)

// Plan is the only billing class Nestplay accepts.
type Plan string

const PlanFree Plan = "free"

// ParsePlan accepts only unpaid labels. Empty means free.
func ParsePlan(raw string) (Plan, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "free", "developer", "personal", "noncommercial", "non-commercial":
		return PlanFree, nil
	case "vip", "light", "startup", "basic", "premium", "pro", "enterprise",
		"commercial", "paid":
		return "", fmt.Errorf("%w: %q", ErrPaidPlan, raw)
	default:
		return "", fmt.Errorf("%w: unknown plan %q", ErrPaidPlan, raw)
	}
}

// Config is local provider policy. Secrets stay in the env file; this struct
// only carries plan flags that must remain free.
type Config struct {
	TMDBPlan             Plan
	OpenSubtitlesPlan    Plan
	OpenSubtitlesDevMode bool
}

// Validate reports paid plans and OpenSubtitles test-quota bypasses.
func (c Config) Validate() error {
	if c.TMDBPlan != PlanFree {
		return fmt.Errorf("tmdb: %w", ErrPaidPlan)
	}
	if c.OpenSubtitlesPlan != PlanFree {
		return fmt.Errorf("opensubtitles: %w", ErrPaidPlan)
	}
	if c.OpenSubtitlesDevMode {
		return ErrDevMode
	}
	return nil
}

// OpenSubtitlesAccount is the subset of a login/status payload Nestplay may
// inspect. Adapters must map provider JSON into this type; they must not make
// download decisions themselves.
type OpenSubtitlesAccount struct {
	VIP                bool
	Level              string
	AllowedDownloads   int
	RemainingDownloads int
	RemainingKnown     bool
}

// ValidateFreeAccount rejects VIP/Pro ranks and allowances above the documented
// free authenticated range (help pages cite 10–20/day; the free consumer table
// is 5/day). Nestplay still self-caps at 5.
func (a OpenSubtitlesAccount) ValidateFreeAccount() error {
	if a.VIP {
		return fmt.Errorf("%w: vip=true", ErrPaidAccount)
	}
	level := strings.ToLower(a.Level)
	switch {
	case strings.Contains(level, "vip"),
		strings.Contains(level, "pro"),
		strings.Contains(level, "enterprise"),
		strings.Contains(level, "premium"):
		return fmt.Errorf("%w: level %q", ErrPaidAccount, a.Level)
	}
	if a.AllowedDownloads > freeOpenSubtitlesAccountCeiling {
		return fmt.Errorf("%w: allowed_downloads=%d", ErrPaidAccount, a.AllowedDownloads)
	}
	return nil
}

// LocalSpanish is already-present dialogue on disk or in the media file.
type LocalSpanish struct {
	EmbeddedFull bool
	SidecarFull  bool
}

func (l LocalSpanish) Present() bool {
	return l.EmbeddedFull || l.SidecarFull
}

// SubtitleCandidate is a provider hit. Hash or confirmed release identity is
// required before spending a free download.
type SubtitleCandidate struct {
	HashMatch         bool
	ReleaseMatch      bool
	MachineTranslated bool
}

func (c SubtitleCandidate) StrongMatch() bool {
	return c.HashMatch || c.ReleaseMatch
}

// Budget counts free downloads in a UTC day. Callers persist it; this type
// only enforces the cap.
type Budget struct {
	Day       time.Time
	Downloads int
}

func (b Budget) remaining(now time.Time) int {
	if !sameUTCDay(b.Day, now.UTC()) {
		return FreeOpenSubtitlesDownloadsPerDay
	}
	left := FreeOpenSubtitlesDownloadsPerDay - b.Downloads
	if left < 0 {
		return 0
	}
	return left
}

func sameUTCDay(a, b time.Time) bool {
	ay, am, ad := a.UTC().Date()
	by, bm, bd := b.UTC().Date()
	return ay == by && am == bm && ad == bd
}

// AuthorizeDownload is the only gate a future OpenSubtitles adapter may use
// before requesting a subtitle file. Searching remains unlimited; downloading
// is not.
func AuthorizeDownload(cfg Config, account OpenSubtitlesAccount, budget Budget, local LocalSpanish, candidate SubtitleCandidate, now time.Time) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := account.ValidateFreeAccount(); err != nil {
		return err
	}
	if local.Present() {
		return ErrLocalSpanish
	}
	if candidate.MachineTranslated {
		return ErrMachineTranslated
	}
	if !candidate.StrongMatch() {
		return ErrWeakMatch
	}
	left := budget.remaining(now)
	if account.RemainingKnown && account.RemainingDownloads < left {
		left = account.RemainingDownloads
	}
	if left < 1 {
		return ErrQuotaExhausted
	}
	return nil
}
