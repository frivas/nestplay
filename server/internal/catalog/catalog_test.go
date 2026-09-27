package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	if items[0].Title != "Pilot" || items[0].Filename != "Pilot.mp4" || items[0].ID == "" || items[0].Collection != "shows" {
		t.Fatalf("nested: %+v", items[0])
	}
	if items[1].Title != "The Guest" || items[1].Filename != "The Guest.mkv" || items[1].Collection != "" {
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

func TestNewestFileAppearsFirst(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "Alpha.mkv"), []byte("x"))
	time.Sleep(20 * time.Millisecond)
	mustWrite(t, filepath.Join(root, "Zebra.mkv"), []byte("x"))
	items, err := List(root)
	if err != nil || len(items) != 2 || items[0].Title != "Zebra" || items[1].Title != "Alpha" {
		t.Fatalf("newest should be first: %+v %v", items, err)
	}

	c, err := open(root, "", 0, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	time.Sleep(20 * time.Millisecond)
	mustWrite(t, filepath.Join(root, "Heat.mkv"), []byte("x"))
	if err := c.Scan(); err != nil {
		t.Fatal(err)
	}
	got := c.Items()
	if len(got) != 3 || got[0].Title != "Heat" {
		t.Fatalf("admitted file should lead: %+v", got)
	}
}

func TestListSkipsPartFiles(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "Heat.mkv.part"), []byte("x"))
	mustWrite(t, filepath.Join(root, "Heat.mkv"), []byte("x"))
	items, err := List(root)
	if err != nil || len(items) != 1 || items[0].Filename != "Heat.mkv" {
		t.Fatalf("got %+v %v", items, err)
	}
}

func TestGrowingFileWaitsToSettle(t *testing.T) {
	root := t.TempDir()
	now := time.Now()
	clock := func() time.Time { return now }
	c, err := open(root, filepath.Join(t.TempDir(), "library.sqlite"), 100*time.Millisecond, clock)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	path := filepath.Join(root, "Heat.mkv")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("abc")); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := c.Scan(); err != nil {
		t.Fatal(err)
	}
	if len(c.Items()) != 0 {
		t.Fatalf("copying file admitted early: %+v", c.Items())
	}

	now = now.Add(40 * time.Millisecond)
	if _, err := f.Write([]byte("def")); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := c.Scan(); err != nil {
		t.Fatal(err)
	}
	if len(c.Items()) != 0 {
		t.Fatalf("still copying: %+v", c.Items())
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	now = now.Add(120 * time.Millisecond)
	if err := c.Scan(); err != nil {
		t.Fatal(err)
	}
	items := c.Items()
	if len(items) != 1 || items[0].Title != "Heat" {
		t.Fatalf("settled: %+v", items)
	}
	file, err := c.FileByID(items[0].ID)
	if err != nil || filepath.Base(file.Path) != "Heat.mkv" {
		t.Fatalf("FileByID: %+v %v", file, err)
	}
}

func TestSnapshotIDsMatchAfterRescan(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "Heat.mkv"), []byte("x"))
	c, err := open(root, filepath.Join(t.TempDir(), "library.sqlite"), 0, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	first := c.Items()
	if err := c.Scan(); err != nil {
		t.Fatal(err)
	}
	again := c.Items()
	if len(first) != 1 || first[0].ID != again[0].ID {
		t.Fatalf("ids diverged: %+v %+v", first, again)
	}
}

func TestSubscribeSignalsWhenCatalogChanges(t *testing.T) {
	root := t.TempDir()
	c, err := open(root, "", 0, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	ch, unsub := c.Subscribe()
	defer unsub()
	mustWrite(t, filepath.Join(root, "Heat.mkv"), []byte("x"))
	if err := c.Scan(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ch:
	default:
		t.Fatal("expected change signal")
	}
}

func TestRewrittenFileStaysVisibleUntilSettled(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Heat.mkv")
	mustWrite(t, path, []byte("abc"))
	now := time.Now()
	clock := func() time.Time { return now }
	c, err := open(root, "", 100*time.Millisecond, clock)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if len(c.Items()) != 1 {
		t.Fatalf("startup: %+v", c.Items())
	}

	if err := os.WriteFile(path, []byte("abcdef"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := c.Scan(); err != nil {
		t.Fatal(err)
	}
	if len(c.Items()) != 1 || c.Items()[0].Title != "Heat" {
		t.Fatalf("rewriting hid the title: %+v", c.Items())
	}

	now = now.Add(120 * time.Millisecond)
	if err := c.Scan(); err != nil {
		t.Fatal(err)
	}
	file, err := c.FileByID(c.Items()[0].ID)
	if err != nil || file.size != 6 {
		t.Fatalf("settled rewrite: %+v %v", file, err)
	}
}

func TestCollectionIsParentFolderNotMacPath(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "Mateo", "PJ Masks", "09 - Gatuno.mp4"), []byte("x"))
	mustWrite(t, filepath.Join(root, "Mateo", "Simon Castellano", "02 - Gasolinera.mp4"), []byte("x"))
	items, err := List(root)
	if err != nil || len(items) != 2 {
		t.Fatalf("got %+v %v", items, err)
	}
	byFile := map[string]Item{}
	for _, item := range items {
		byFile[item.Filename] = item
	}
	pj := byFile["09 - Gatuno.mp4"]
	simon := byFile["02 - Gasolinera.mp4"]
	if pj.Collection != "Mateo/PJ Masks" || simon.Collection != "Mateo/Simon Castellano" {
		t.Fatalf("collections: %+v", items)
	}
	if strings.Contains(pj.Collection, root) || strings.Contains(simon.Collection, root) {
		t.Fatalf("mac path leaked into collection: %+v", items)
	}
}

func TestSetVisibleHidesUnplayableFromItems(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "play.mp4"), []byte("a"))
	mustWrite(t, filepath.Join(root, "hide.mp4"), []byte("b"))
	cat, err := OpenImmediate(root, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cat.Close() })
	cat.SetVisible(func(path string) bool {
		return filepath.Base(path) == "play.mp4"
	})
	items := cat.Items()
	if len(items) != 1 || items[0].Filename != "play.mp4" {
		t.Fatalf("%+v", items)
	}
	if _, err := cat.FileByID(items[0].ID); err != nil {
		t.Fatal(err)
	}
	hidden := cat.Hidden()
	if len(hidden) != 1 || hidden[0].Item.Filename != "hide.mp4" {
		t.Fatalf("%+v", hidden)
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
