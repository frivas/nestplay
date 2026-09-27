package downloads

import (
	"errors"
	"testing"
)

func TestPickPrefers1080H264Over4KAV1(t *testing.T) {
	got, err := Pick([]Option{
		{ID: "av1", Height: 2160, VCodec: "av01.0"},
		{ID: "avc", Height: 1080, VCodec: "avc1.640028"},
		{ID: "low", Height: 720, VCodec: "avc1.64001f"},
	})
	if err != nil || got.ID != "avc" {
		t.Fatalf("got %+v %v", got, err)
	}
}

func TestPickTakesHighestWhenNoH264At1080(t *testing.T) {
	got, err := Pick([]Option{
		{ID: "av1", Height: 2160, VCodec: "av01.0"},
		{ID: "vp9", Height: 1440, VCodec: "vp9"},
		{ID: "low", Height: 720, VCodec: "avc1.64001f"},
	})
	if err != nil || got.ID != "av1" {
		t.Fatalf("got %+v %v", got, err)
	}
}

func TestPickAllows720WhenBest(t *testing.T) {
	got, err := Pick([]Option{
		{ID: "p720", Height: 720, VCodec: "avc1.64001f"},
		{ID: "p360", Height: 360, VCodec: "avc1.4d401e"},
	})
	if err != nil || got.ID != "p720" {
		t.Fatalf("got %+v %v", got, err)
	}
}

func TestPickRejectsBelow720(t *testing.T) {
	_, err := Pick([]Option{
		{ID: "p360", Height: 360, VCodec: "avc1.4d401e"},
		{ID: "p240", Height: 240, VCodec: "av01.0"},
	})
	if !errors.Is(err, ErrBelow720) {
		t.Fatalf("got %v", err)
	}
}
