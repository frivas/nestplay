package downloads

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"nestplay.local/server/internal/catalog"
	"nestplay.local/server/internal/storage"
)

func TestConvertHiddenReplacesOriginalAfterSuccess(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "clip.mkv")
	if err := os.WriteFile(src, []byte("av1"), 0o600); err != nil {
		t.Fatal(err)
	}
	cat, err := catalog.OpenImmediate(root, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cat.Close() })
	cat.SetVisible(func(string) bool { return false })
	if err := ConvertHidden(context.Background(), cat, stubConvert{}, func(string) (uint64, error) {
		return storage.Reserve + 5, nil
	}, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("original still present: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(root, "clip.mp4")); err != nil || string(data) != "playable" {
		t.Fatalf("dest %v %q", err, data)
	}
	if _, err := os.Stat(filepath.Join(root, "clip.mp4.part")); !os.IsNotExist(err) {
		t.Fatal("left a part file")
	}
}

func TestConvertHiddenKeepsOriginalWhenReserveRefuses(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "clip.mkv")
	if err := os.WriteFile(src, []byte("av1"), 0o600); err != nil {
		t.Fatal(err)
	}
	cat, err := catalog.OpenImmediate(root, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cat.Close() })
	cat.SetVisible(func(string) bool { return false })
	err = ConvertHidden(context.Background(), cat, stubConvert{}, func(string) (uint64, error) {
		return storage.Reserve, nil
	}, t.TempDir())
	if !errors.Is(err, storage.ErrRefused) {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatal("original was removed")
	}
}

func TestConvertHiddenFailureLeavesOriginalAndNoPart(t *testing.T) {
	root := t.TempDir()
	work := t.TempDir()
	src := filepath.Join(root, "clip.mkv")
	if err := os.WriteFile(src, []byte("av1"), 0o600); err != nil {
		t.Fatal(err)
	}
	cat, err := catalog.OpenImmediate(root, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cat.Close() })
	cat.SetVisible(func(string) bool { return false })
	if err := ConvertHidden(context.Background(), cat, failConvert{}, func(string) (uint64, error) {
		return storage.Reserve + 5, nil
	}, work); err == nil {
		t.Fatal("expected convert error")
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatal("original was removed")
	}
	if _, err := os.Stat(filepath.Join(root, "clip.mp4")); !os.IsNotExist(err) {
		t.Fatal("wrote a dest file")
	}
	if leftovers(t, work) {
		t.Fatal("left work files")
	}
}

func TestSweepRemovesStaleTemps(t *testing.T) {
	root := t.TempDir()
	work := filepath.Join(root, ".nestplay-work")
	if err := os.MkdirAll(filepath.Join(work, "yt-old"), 0o700); err != nil {
		t.Fatal(err)
	}
	part := filepath.Join(root, "clip.mp4.part")
	if err := os.WriteFile(part, []byte("partial"), 0o600); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(root, "Heat.mp4")
	if err := os.WriteFile(keep, []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	Sweep(root, work)
	if _, err := os.Stat(work); !os.IsNotExist(err) {
		t.Fatal("work root remained")
	}
	if _, err := os.Stat(part); !os.IsNotExist(err) {
		t.Fatal("part remained")
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatal("deleted a library file")
	}
}
