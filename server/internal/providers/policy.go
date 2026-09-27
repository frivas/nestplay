package providers

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Unpaid OpenSubtitles.com consumer. Nestplay never buys Light–Enterprise, VIP,
// or Pro. The owner's consumer has "Allow anonymous downloads" and "Under
// dev" checked: Api-Key only, no JWT, 100 downloads/24h. Search is unlimited.
// Getting started: https://opensubtitles.tawk.help/article/getting-started
// About the API: https://opensubtitles.tawk.help/article/about-the-api
const (
	FreeOpenSubtitlesAnonymousDownloadsPerDay = 5
	FreeOpenSubtitlesLoggedInDownloadsPerDay  = 20
	FreeOpenSubtitlesDevDownloadsPerDay       = 100
	FreeOpenSubtitlesRequestsPerSec           = 5
	OpenSubtitlesAPIHost                      = "api.opensubtitles.com"
	OpenSubtitlesUserAgent                    = "Nestplay v0.1.0"
)

var (
	ErrPaidPlan          = errors.New("nestplay PoC allows only unpaid provider plans")
	ErrPaidAccount       = errors.New("refusing a paid, VIP, or Pro OpenSubtitles account")
	ErrVIPEndpoint       = errors.New("refusing OpenSubtitles VIP API host")
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
	TMDBPlan               Plan
	OpenSubtitlesPlan      Plan
	OpenSubtitlesDevMode   bool
	OpenSubtitlesAnonymous bool
}

func defaultConfig() Config {
	return Config{
		TMDBPlan:               PlanFree,
		OpenSubtitlesPlan:      PlanFree,
		OpenSubtitlesDevMode:   true,
		OpenSubtitlesAnonymous: true,
	}
}

// Validate reports paid plans. Under-dev and anonymous downloads are unpaid
// consumer flags, not paid packages.
func (c Config) Validate() error {
	if c.TMDBPlan != PlanFree {
		return fmt.Errorf("tmdb: %w", ErrPaidPlan)
	}
	if c.OpenSubtitlesPlan != PlanFree {
		return fmt.Errorf("opensubtitles: %w", ErrPaidPlan)
	}
	return nil
}

// AllowsAnonymousDownload is true when the consumer can hit /download with
// Api-Key and User-Agent only. JWT login is skipped: login is rate-limited
// and 401s must not be retried with the same credentials.
func (c Config) AllowsAnonymousDownload() bool {
	return c.OpenSubtitlesAnonymous || c.OpenSubtitlesDevMode
}

// DownloadCeiling is Nestplay's self-cap for one UTC day.
// Under-dev: 100 without user auth. Otherwise anonymous 5/IP or free logged-in 20.
func (c Config) DownloadCeiling() int {
	switch {
	case c.OpenSubtitlesDevMode:
		return FreeOpenSubtitlesDevDownloadsPerDay
	case !c.OpenSubtitlesAnonymous:
		return FreeOpenSubtitlesLoggedInDownloadsPerDay
	default:
		return FreeOpenSubtitlesAnonymousDownloadsPerDay
	}
}

// CheckAPIHost refuses vip-api.opensubtitles.com. Adapters must keep using
// api.opensubtitles.com and follow redirects. If login ever returns a VIP
// base_url, stop rather than sending a JWT there.
func CheckAPIHost(host string) error {
	cleaned := strings.ToLower(strings.TrimSpace(host))
	cleaned = strings.TrimPrefix(cleaned, "https://")
	cleaned = strings.TrimPrefix(cleaned, "http://")
	cleaned = strings.TrimSuffix(cleaned, "/")
	if cleaned == "" {
		return nil
	}
	if strings.Contains(cleaned, "vip-api") {
		return fmt.Errorf("%w: %s", ErrVIPEndpoint, host)
	}
	return nil
}

// EncodeQuery follows OpenSubtitles best practice: sorted keys, + for spaces.
func EncodeQuery(params map[string]string) string {
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	encoded := values.Encode()
	return strings.ReplaceAll(encoded, "%20", "+")
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

// ValidateFreeAccount rejects VIP/Pro ranks and allowances above this
// consumer's unpaid ceiling. Under-dev may report 100; 1000 is VIP.
func (a OpenSubtitlesAccount) ValidateFreeAccount(cfg Config) error {
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
	if a.AllowedDownloads > cfg.DownloadCeiling() {
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
// required before spending a free download. Searching does not spend quota.
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

func (b Budget) remaining(limit int, now time.Time) int {
	if !sameUTCDay(b.Day, now.UTC()) {
		return limit
	}
	left := limit - b.Downloads
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
// is not. Under-dev quota is shared on the consumer key, so weak matches still
// must not spend it.
func AuthorizeDownload(cfg Config, account OpenSubtitlesAccount, budget Budget, local LocalSpanish, candidate SubtitleCandidate, now time.Time) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := account.ValidateFreeAccount(cfg); err != nil {
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
	left := budget.remaining(cfg.DownloadCeiling(), now)
	if account.RemainingKnown && account.RemainingDownloads < left {
		left = account.RemainingDownloads
	}
	if left < 1 {
		return ErrQuotaExhausted
	}
	return nil
}
