package artwork

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"nestplay.local/server/internal/tools"
)

const (
	width  = 640
	height = 360
)

// JPEG returns a 16:9 poster for a watched-folder file. It prefers a cached
// ffmpeg still and falls back to a designed placeholder. src is a Mac path
// and must never appear in the bytes or HTTP response.
func JPEG(id, src string) ([]byte, error) {
	info, err := os.Stat(src)
	if err != nil {
		return placeholder(), nil
	}
	cachePath, err := cacheFile(id, info.ModTime())
	if err != nil {
		return placeholder(), nil
	}
	if data, err := os.ReadFile(cachePath); err == nil && jpegSOI(data) { // #nosec G304 -- cache path is generated from a hex Nestplay id
		return data, nil
	}
	if err := extract(src, cachePath); err != nil {
		return placeholder(), nil
	}
	data, err := os.ReadFile(cachePath) // #nosec G304 -- cache path is generated from a hex Nestplay id
	if err != nil || !jpegSOI(data) {
		return placeholder(), nil
	}
	return data, nil
}

func extract(src, dest string) error {
	ffmpeg, err := tools.FFmpeg()
	if err != nil {
		return err
	}
	ss := strconv.FormatFloat(seekSeconds(durationSeconds(src)), 'f', 3, 64)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	// src and dest are catalog/cache paths owned by Nestplay, never request input.
	cmd := exec.CommandContext(ctx, ffmpeg, // #nosec G204
		"-hide_banner", "-loglevel", "error", "-nostdin",
		"-ss", ss,
		"-i", src,
		"-frames:v", "1",
		"-vf", "scale=640:360:force_original_aspect_ratio=increase,crop=640:360",
		"-q:v", "4",
		"-y", dest,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w", err)
	}
	return nil
}

func durationSeconds(src string) float64 {
	ffprobe, err := tools.FFprobe()
	if err != nil {
		return 0
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, ffprobe, // #nosec G204
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		src,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0
	}
	return seconds
}

func seekSeconds(duration float64) float64 {
	if duration <= 0 {
		return 45
	}
	if duration < 50 {
		return duration * 0.35
	}
	seconds := duration * 0.18
	if seconds < 45 {
		seconds = 45
	}
	if seconds > 180 {
		seconds = 180
	}
	if seconds > duration-1 {
		return duration * 0.4
	}
	return seconds
}

func cacheFile(id string, mod time.Time) (string, error) {
	root, err := cacheRoot()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", err
	}
	name := id + "-" + strconv.FormatInt(mod.Unix(), 10) + "-v3.jpg"
	return filepath.Join(root, name), nil
}

func cacheRoot() (string, error) {
	if dir := os.Getenv("NESTPLAY_ARTWORK_CACHE"); dir != "" {
		return dir, nil
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "nestplay", "artwork"), nil
}

func jpegSOI(data []byte) bool {
	return len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8
}

func Placeholder() []byte {
	return placeholder()
}

func placeholder() []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	top := color.RGBA{R: 0x1B, G: 0x41, B: 0x48, A: 0xFF}
	bottom := color.RGBA{R: 0x0B, G: 0x1E, B: 0x22, A: 0xFF}
	for y := 0; y < height; y++ {
		t := float64(y) / float64(height-1)
		c := color.RGBA{
			R: mix(top.R, bottom.R, t),
			G: mix(top.G, bottom.G, t),
			B: mix(top.B, bottom.B, t),
			A: 0xFF,
		}
		draw.Draw(img, image.Rect(0, y, width, y+1), &image.Uniform{C: c}, image.Point{}, draw.Src)
	}
	cream := color.RGBA{R: 0xF6, G: 0xEE, B: 0xD8, A: 0xFF}
	gold := color.RGBA{R: 0xE8, G: 0xC8, B: 0x5A, A: 0xFF}
	cx, cy := width/2, height/2
	drawPlay(img, cx, cy, 54, gold)
	drawPlay(img, cx, cy, 42, cream)
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	return buf.Bytes()
}

func mix(a, b uint8, t float64) uint8 {
	return uint8(float64(a)*(1-t) + float64(b)*t)
}

func drawPlay(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	for y := -size; y <= size; y++ {
		span := (size - abs(y)) * 5 / 6
		for x := 0; x <= span; x++ {
			img.SetRGBA(cx-size/4+x, cy+y, c)
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
