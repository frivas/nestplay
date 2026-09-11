package providers

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvDefaultsToFreeAndRejectsPaid(t *testing.T) {
	cfg, err := parseEnv("# comment\nTMDB_API_KEY=secret\n")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TMDBPlan != PlanFree || cfg.OpenSubtitlesPlan != PlanFree || !cfg.OpenSubtitlesDevMode || !cfg.OpenSubtitlesAnonymous {
		t.Fatalf("%+v", cfg)
	}
	if _, err := parseEnv("OPENSUBTITLES_PLAN=vip\n"); !errors.Is(err, ErrPaidPlan) {
		t.Fatalf("vip: %v", err)
	}
	dev, err := parseEnv("OPENSUBTITLES_DEV_MODE=true\nOPENSUBTITLES_ANONYMOUS=true\n")
	if err != nil {
		t.Fatal(err)
	}
	if err := dev.Validate(); err != nil {
		t.Fatalf("dev validate: %v", err)
	}
	off, err := parseEnv("OPENSUBTITLES_DEV_MODE=false\nOPENSUBTITLES_ANONYMOUS=false\n")
	if err != nil {
		t.Fatal(err)
	}
	if off.OpenSubtitlesDevMode || off.OpenSubtitlesAnonymous {
		t.Fatalf("explicit false: %+v", off)
	}
}

func TestLoadOptionalConfigMissingIsOKPaidFileFails(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("NESTPLAY_PROVIDERS_FILE", "")
	t.Setenv("HOME", dir)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if _, path, err := LoadOptionalConfig(); err != nil || path != "" {
		t.Fatalf("missing: path=%q err=%v", path, err)
	}

	missing := filepath.Join(dir, "missing.env")
	t.Setenv("NESTPLAY_PROVIDERS_FILE", missing)
	if _, _, err := LoadOptionalConfig(); err == nil {
		t.Fatal("explicit missing path should fail")
	}

	paid := filepath.Join(dir, "providers.env")
	if err := os.WriteFile(paid, []byte("OPENSUBTITLES_PLAN=pro\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NESTPLAY_PROVIDERS_FILE", paid)
	if _, _, err := LoadOptionalConfig(); !errors.Is(err, ErrPaidPlan) {
		t.Fatalf("paid file: %v", err)
	}

	free := filepath.Join(dir, "free.env")
	if err := os.WriteFile(free, []byte("TMDB_PLAN=free\nOPENSUBTITLES_PLAN=free\nOPENSUBTITLES_DEV_MODE=true\nOPENSUBTITLES_ANONYMOUS=true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NESTPLAY_PROVIDERS_FILE", free)
	cfg, path, err := LoadOptionalConfig()
	if err != nil || path != free || cfg.Validate() != nil {
		t.Fatalf("free file: path=%q err=%v cfg=%+v", path, err, cfg)
	}
}
