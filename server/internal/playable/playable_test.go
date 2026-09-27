package playable

import "testing"

func TestClassifyShowH264MP4(t *testing.T) {
	if Classify("h264", "mov,mp4,m4a", "yuv420p") != Show {
		t.Fatal("h264 mp4 must show")
	}
	if Classify("avc1", "isom,mp4", "yuv420p") != Show {
		t.Fatal("avc1 mp4 must show")
	}
}

func TestClassifyHideUnplayable(t *testing.T) {
	cases := []struct{ codec, format, pix string }{
		{"av1", "mov,mp4", "yuv420p"},
		{"vp9", "webm", "yuv420p"},
		{"hevc", "mov,mp4", "yuv420p"},
		{"h264", "matroska,webm", "yuv420p"},
		{"h264", "mov,mp4", "yuv420p10le"},
		{"", "mov,mp4", "yuv420p"},
	}
	for _, tc := range cases {
		if Classify(tc.codec, tc.format, tc.pix) != Hide {
			t.Fatalf("expected hide %+v", tc)
		}
	}
}
