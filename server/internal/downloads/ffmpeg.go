package downloads

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"nestplay.local/server/internal/tools"
)

// FFmpeg writes a Living Room H.264+AAC MP4. It does not upscale.
type FFmpeg struct {
	Bin string
}

func (f FFmpeg) MakePlayable(ctx context.Context, src, dest string) error {
	bin := f.Bin
	if bin == "" {
		var err error
		bin, err = tools.FFmpeg()
		if err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return err
	}
	args := []string{
		"-hide_banner", "-loglevel", "error", "-nostdin", "-y",
		"-i", src,
		"-map", "0:v:0", "-map", "0:a:0?",
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "20", "-pix_fmt", "yuv420p",
		"-vf", "scale='min(1920,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease",
		"-c:a", "aac", "-b:a", "192k",
		"-movflags", "+faststart",
		"-f", "mp4",
		dest,
	}
	cmd := exec.CommandContext(ctx, bin, args...) // #nosec G204 -- src/dest are Nestplay-owned work paths
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w", err)
	}
	return nil
}
