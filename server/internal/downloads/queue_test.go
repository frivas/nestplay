package downloads

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"nestplay.local/server/internal/storage"
)

type stubFetch struct {
	err error
}

func (s stubFetch) Fetch(_ context.Context, _, destDir string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	path := filepath.Join(destDir, "source.mp4")
	if err := os.WriteFile(path, []byte("src"), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

type stubConvert struct{}

func (stubConvert) MakePlayable(_ context.Context, _, dest string) error {
	return os.WriteFile(dest, []byte("playable"), 0o600)
}

func TestProcessOnePublishesPlayableFile(t *testing.T) {
	root := t.TempDir()
	work := t.TempDir()
	q := Open(root, work, stubFetch{}, stubConvert{}, func(string) (uint64, error) {
		return storage.Reserve + 10, nil
	})
	q.Enqueue([]Item{{ID: "dQw4w9WgXcQ", Title: "Heat"}})
	if err := q.ProcessOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	list := q.List()
	if len(list) != 1 || list[0].State != StateComplete {
		t.Fatalf("%+v", list)
	}
	if _, err := os.Stat(filepath.Join(root, "Heat-dQw4w9WgXcQ.mp4")); err != nil {
		t.Fatal(err)
	}
	if leftovers(t, work) {
		t.Fatal("left work files")
	}
}

func TestProcessOneTooLowStaysOffDisk(t *testing.T) {
	root := t.TempDir()
	q := Open(root, t.TempDir(), stubFetch{err: ErrBelow720}, stubConvert{}, func(string) (uint64, error) {
		return storage.Reserve + 10, nil
	})
	q.Enqueue([]Item{{ID: "dQw4w9WgXcQ", Title: "Tiny"}})
	if err := q.ProcessOne(context.Background()); !errors.Is(err, ErrBelow720) {
		t.Fatalf("got %v", err)
	}
	if q.List()[0].State != StateTooLow {
		t.Fatalf("%+v", q.List())
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("wrote %+v", entries)
	}
}

func TestProcessOneWaitsWhenReserveRefuses(t *testing.T) {
	root := t.TempDir()
	q := Open(root, t.TempDir(), stubFetch{}, stubConvert{}, func(string) (uint64, error) {
		return storage.Reserve, nil
	})
	q.Enqueue([]Item{{ID: "dQw4w9WgXcQ", Title: "Heat"}})
	if err := q.ProcessOne(context.Background()); !errors.Is(err, storage.ErrRefused) {
		t.Fatalf("got %v", err)
	}
	if q.List()[0].State != StateWaitingSpace {
		t.Fatalf("%+v", q.List())
	}
}

func TestProcessOneCreatesMissingWorkRoot(t *testing.T) {
	root := t.TempDir()
	work := filepath.Join(t.TempDir(), "missing", "work")
	q := Open(root, work, stubFetch{}, stubConvert{}, func(string) (uint64, error) {
		return storage.Reserve + 10, nil
	})
	q.Enqueue([]Item{{ID: "dQw4w9WgXcQ", Title: "Heat"}})
	if err := q.ProcessOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	if q.List()[0].State != StateComplete {
		t.Fatalf("%+v", q.List())
	}
}

func TestProcessOneConvertFailureLeavesNoDest(t *testing.T) {
	root := t.TempDir()
	work := t.TempDir()
	q := Open(root, work, stubFetch{}, failConvert{}, func(string) (uint64, error) {
		return storage.Reserve + 10, nil
	})
	q.Enqueue([]Item{{ID: "dQw4w9WgXcQ", Title: "Heat"}})
	if err := q.ProcessOne(context.Background()); err == nil {
		t.Fatal("expected convert error")
	}
	if q.List()[0].State != StateFailed {
		t.Fatalf("%+v", q.List())
	}
	if leftovers(t, root) {
		t.Fatal("left files in the watched folder")
	}
	if leftovers(t, work) {
		t.Fatal("left work files")
	}
}

func leftovers(t *testing.T, dir string) bool {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	return len(entries) > 0
}

type failConvert struct{}

func (failConvert) MakePlayable(context.Context, string, string) error {
	return errors.New("convert failed")
}

func TestEnqueueIgnoresBadIDsAndDupes(t *testing.T) {
	q := Open(t.TempDir(), t.TempDir(), stubFetch{}, stubConvert{}, nil)
	q.Enqueue([]Item{{ID: "bad", Title: "x"}, {ID: "dQw4w9WgXcQ", Title: "A"}})
	q.Enqueue([]Item{{ID: "dQw4w9WgXcQ", Title: "B"}})
	if len(q.List()) != 1 || q.List()[0].Title != "A" {
		t.Fatalf("%+v", q.List())
	}
}
