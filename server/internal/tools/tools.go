package tools

import (
	"os"
	"os/exec"
	"path/filepath"
)

const (
	DefaultYTDLP     = "/opt/homebrew/Caskroom/miniconda/base/envs/ytdlp/bin/yt-dlp"
	DefaultFFmpegDir = "/opt/homebrew/bin"
)

var (
	ytdlp  string
	ffmpeg string
)

// Configure sets the yt-dlp file and ffmpeg directory or binary.
// Empty values use the Nestplay Mac defaults when those files exist,
// otherwise PATH.
func Configure(ytdlpPath, ffmpegPath string) {
	ytdlp = ytdlpPath
	ffmpeg = ffmpegPath
}

// YTDLP returns the yt-dlp executable.
func YTDLP() (string, error) {
	return resolve("yt-dlp", first(ytdlp, DefaultYTDLP))
}

// FFmpeg returns the ffmpeg executable.
func FFmpeg() (string, error) {
	return resolve("ffmpeg", first(ffmpeg, DefaultFFmpegDir))
}

// FFprobe returns the ffprobe executable (same directory as ffmpeg).
func FFprobe() (string, error) {
	return resolve("ffprobe", first(ffmpeg, DefaultFFmpegDir))
}

func first(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func resolve(name, configured string) (string, error) {
	if configured != "" {
		info, err := os.Stat(configured)
		if err == nil {
			if info.IsDir() {
				candidate := filepath.Join(configured, name)
				if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
					return candidate, nil
				}
			} else if name == "yt-dlp" || filepath.Base(configured) == name {
				return configured, nil
			} else {
				dir := filepath.Dir(configured)
				candidate := filepath.Join(dir, name)
				if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
					return candidate, nil
				}
			}
		}
	}
	return exec.LookPath(name)
}
