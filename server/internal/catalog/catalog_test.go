package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListEmptyAndMissingRoots(t *testing.T) {
	items, err := List("")
	if err != nil || len(items) != 0 {
		t.Fatalf("empty root: %+v %v", items, err)
	}
	items, err = List(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil || len(items) != 0 {
		t.Fatalf("missing root: %+v %v", items, err)
	}
}

func TestListFindsVideosAndSkipsOthers(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "The Guest.mkv"), []byte("x"))
	mustWrite(t, filepath.Join(root, "notes.txt"), []byte("no"))
	mustWrite(t, filepath.Join(root, ".hidden.mp4"), []byte("no"))
	mustWrite(t, filepath.Join(root, "shows", "Pilot.mp4"), []byte("x"))
	mustWrite(t, filepath.Join(root, "not-a-dir"), []byte("x"))

	items, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %+v", items)
	}
	if items[0].Title != "Pilot" || items[0].Filename != "Pilot.mp4" || items[0].ID == "" {
		t.Fatalf("nested: %+v", items[0])
	}
	if items[1].Title != "The Guest" || items[1].Filename != "The Guest.mkv" {
		t.Fatalf("root file: %+v", items[1])
	}
	file, err := FileForID(root, items[1].ID)
	if err != nil || file.Item.ID != items[1].ID || filepath.Base(file.Path) != "The Guest.mkv" {
		t.Fatalf("FileForID: %+v %v", file, err)
	}
	if _, err := FileForID(root, "missingmissingmissingmissingmiss"); err != ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	again, err := List(root)
	if err != nil || again[0].ID != items[0].ID || again[1].ID != items[1].ID {
		t.Fatalf("ids must be stable: %+v %+v", items, again)
	}
}

func TestListFileRootIsEmptyCatalog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.mp4")
	mustWrite(t, path, []byte("x"))
	items, err := List(path)
	if err != nil || len(items) != 0 {
		t.Fatalf("file root: %+v %v", items, err)
	}
}

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
