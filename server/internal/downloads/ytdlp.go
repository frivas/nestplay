package downloads

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"nestplay.local/server/internal/tools"
	"nestplay.local/server/internal/youtube"
)

// YTDLP fetches with yt-dlp. Cookies stay on the Mac file, never in logs.
type YTDLP struct {
	Cookies string
	Bin     string
}

func (y YTDLP) Fetch(ctx context.Context, id, destDir string) (string, error) {
	if !youtube.VideoID(id) {
		return "", errors.New("invalid id")
	}
	bin := y.Bin
	if bin == "" {
		var err error
		bin, err = tools.YTDLP()
		if err != nil {
			return "", err
		}
	}
	outTmpl := filepath.Join(destDir, "source.%(ext)s")
	args := []string{
		"--no-progress", "--no-playlist", "--no-warnings",
		"--merge-output-format", "mp4",
		"--format", "bestvideo[height>=1080][vcodec^=avc1]+bestaudio/bestvideo[height>=1080]+bestaudio/bestvideo[height>=720][vcodec^=avc1]+bestaudio/bestvideo[height>=720]+bestaudio/best[height>=720]",
		"-o", outTmpl,
		"--", "https://www.youtube.com/watch?v=" + id,
	}
	if strings.TrimSpace(y.Cookies) != "" {
		if _, err := os.Stat(y.Cookies); err == nil {
			args = append([]string{"--cookies", y.Cookies}, args...)
		}
	}
	cmd := exec.CommandContext(ctx, bin, args...) // #nosec G204 -- bin is LookPath yt-dlp; id is a validated YouTube video id
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = nil
	if err := cmd.Run(); err != nil {
		msg := strings.ToLower(stderr.String())
		if strings.Contains(msg, "requested format is not available") || strings.Contains(msg, "only images") {
			return "", ErrBelow720
		}
		return "", fmt.Errorf("yt-dlp: %w", err)
	}
	entries, err := os.ReadDir(destDir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "source.") && !strings.HasSuffix(name, ".part") {
			return filepath.Join(destDir, name), nil
		}
	}
	return "", errors.New("yt-dlp wrote no file")
}
