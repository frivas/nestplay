package providers

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const envFileName = "providers.env"

// SecretsFileCandidates are local-only paths. None of them belong in git.
func SecretsFileCandidates() []string {
	var paths []string
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, "secrets", envFileName))
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "nestplay", envFileName))
	}
	return paths
}

// LoadOptionalConfig reads the first existing secrets file and validates that
// it is an unpaid plan. A missing file is OK: the health server does not need
// providers. NESTPLAY_PROVIDERS_FILE, when set, is the only path consulted and
// must exist. A present file with a paid plan is fatal.
func LoadOptionalConfig() (Config, string, error) {
	if override := strings.TrimSpace(os.Getenv("NESTPLAY_PROVIDERS_FILE")); override != "" {
		return loadFile(override, true)
	}
	for _, path := range SecretsFileCandidates() {
		loaded, used, err := loadFile(path, false)
		if err != nil {
			return Config{}, path, err
		}
		if used != "" {
			return loaded, used, nil
		}
	}
	return defaultConfig(), "", nil
}

func loadFile(path string, required bool) (Config, string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		if required {
			return Config{}, path, fmt.Errorf("provider secrets file not found: %s", path)
		}
		return Config{}, "", nil
	}
	if err != nil {
		return Config{}, path, err
	}
	parsed, err := parseEnv(string(data))
	if err != nil {
		return Config{}, path, fmt.Errorf("%s: %w", path, err)
	}
	if err := parsed.Validate(); err != nil {
		return Config{}, path, fmt.Errorf("%s: %w", path, err)
	}
	return parsed, path, nil
}

func parseEnv(contents string) (Config, error) {
	cfg := defaultConfig()
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Config{}, fmt.Errorf("invalid env line")
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		switch key {
		case "TMDB_PLAN":
			plan, err := ParsePlan(value)
			if err != nil {
				return Config{}, fmt.Errorf("TMDB_PLAN: %w", err)
			}
			cfg.TMDBPlan = plan
		case "OPENSUBTITLES_PLAN":
			plan, err := ParsePlan(value)
			if err != nil {
				return Config{}, fmt.Errorf("OPENSUBTITLES_PLAN: %w", err)
			}
			cfg.OpenSubtitlesPlan = plan
		case "OPENSUBTITLES_DEV_MODE":
			flag, err := parseBool(value)
			if err != nil {
				return Config{}, fmt.Errorf("OPENSUBTITLES_DEV_MODE: %w", err)
			}
			cfg.OpenSubtitlesDevMode = flag
		case "OPENSUBTITLES_ANONYMOUS":
			flag, err := parseBool(value)
			if err != nil {
				return Config{}, fmt.Errorf("OPENSUBTITLES_ANONYMOUS: %w", err)
			}
			cfg.OpenSubtitlesAnonymous = flag
		}
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "1", "true", "yes":
		return true, nil
	case "0", "false", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean %q", value)
	}
}
