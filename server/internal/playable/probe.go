package playable

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"

	"nestplay.local/server/internal/tools"
)

type probeFormat struct {
	FormatName string `json:"format_name"`
}

type probeStream struct {
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	PixFmt    string `json:"pix_fmt"`
}

type probeOut struct {
	Format  probeFormat   `json:"format"`
	Streams []probeStream `json:"streams"`
}

// Probe classifies a Mac file. ffprobe missing or unreadable is Hide.
func Probe(path string) Kind {
	ffprobe, err := tools.FFprobe()
	if err != nil {
		return Hide
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, ffprobe, // #nosec G204 -- path is a catalog file Nestplay already owns
		"-v", "error",
		"-show_entries", "stream=codec_type,codec_name,pix_fmt:format=format_name",
		"-of", "json",
		path,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return Hide
	}
	var out probeOut
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return Hide
	}
	codec, pix := "", ""
	for _, stream := range out.Streams {
		if strings.EqualFold(stream.CodecType, "video") {
			codec = stream.CodecName
			pix = stream.PixFmt
			break
		}
	}
	return Classify(codec, out.Format.FormatName, pix)
}
