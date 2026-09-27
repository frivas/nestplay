package playable

import (
	"strings"
)

// Kind is whether Living Room should show a file or hide it for conversion.
type Kind int

const (
	Show Kind = iota
	Hide
)

// Classify applies the locked Living Room probe. H.264 in MP4/MOV is shown
// (AAC or Opus). AV1, VP9, HEVC, HDR/10-bit, MKV/WebM, and unknown are hidden.
func Classify(videoCodec, formatName, pixFmt string) Kind {
	codec := strings.ToLower(strings.TrimSpace(videoCodec))
	format := strings.ToLower(formatName)
	pix := strings.ToLower(pixFmt)
	if strings.Contains(pix, "10") || strings.Contains(pix, "12") || strings.Contains(pix, "hdr") {
		return Hide
	}
	if strings.Contains(format, "matroska") || strings.Contains(format, "webm") {
		return Hide
	}
	switch codec {
	case "h264", "avc", "avc1":
		if strings.Contains(format, "mp4") || strings.Contains(format, "mov") || strings.Contains(format, "isom") {
			return Show
		}
		return Hide
	default:
		return Hide
	}
}

// H264 reports whether a YouTube vcodec tag is H.264.
func H264(vcodec string) bool {
	v := strings.ToLower(vcodec)
	return strings.HasPrefix(v, "avc") || strings.Contains(v, "h264")
}
