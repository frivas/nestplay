package downloads

import (
	"errors"
	"strings"
)

// ErrBelow720 means YouTube has no rendition at 720p or higher.
var ErrBelow720 = errors.New("no rendition at 720p or higher")

// Option is a YouTube media rendition. IDs stay inside this package.
type Option struct {
	ID     string
	Height int
	VCodec string
}

// Pick chooses a source: never below 720p; never choose below 1080p when
// 1080p or higher exists; prefer H.264 when it already meets the floor.
func Pick(opts []Option) (Option, error) {
	var hi, mid []Option
	for _, opt := range opts {
		if opt.Height >= 1080 {
			hi = append(hi, opt)
		} else if opt.Height >= 720 {
			mid = append(mid, opt)
		}
	}
	if len(hi) > 0 {
		if avc := h264Only(hi); len(avc) > 0 {
			return lowest(avc), nil
		}
		return highest(hi), nil
	}
	if len(mid) == 0 {
		return Option{}, ErrBelow720
	}
	if avc := h264Only(mid); len(avc) > 0 {
		return highest(avc), nil
	}
	return highest(mid), nil
}

func h264Only(opts []Option) []Option {
	var out []Option
	for _, opt := range opts {
		v := strings.ToLower(opt.VCodec)
		if strings.HasPrefix(v, "avc") || strings.Contains(v, "h264") {
			out = append(out, opt)
		}
	}
	return out
}

func highest(opts []Option) Option {
	best := opts[0]
	for _, opt := range opts[1:] {
		if opt.Height > best.Height {
			best = opt
		}
	}
	return best
}

func lowest(opts []Option) Option {
	best := opts[0]
	for _, opt := range opts[1:] {
		if opt.Height < best.Height {
			best = opt
		}
	}
	return best
}
