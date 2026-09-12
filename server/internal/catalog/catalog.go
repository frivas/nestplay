package catalog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

const maxDepth = 6

var videoExt = map[string]struct{}{
	".mp4": {}, ".mkv": {}, ".m4v": {}, ".mov": {}, ".avi": {}, ".webm": {},
}

// ErrNotFound means no watched-folder file owns that Nestplay id.
var ErrNotFound = errors.New("catalog item not found")

// Item is a Nestplay-owned catalog record. It is not a TMDB or filesystem
// schema: the Apple TV only sees these fields.
type Item struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Filename   string `json:"filename"`
	Collection string `json:"collection"`
}

// File is a Mac-side catalog record. Path must never be encoded on the wire.
type File struct {
	Item  Item
	Path  string
	rel   string
	size  int64
	mtime int64
	added int64
}

type pendingFile struct {
	file        File
	size        int64
	stableSince time.Time
}

// Catalog is the in-memory snapshot of the watched folder. List and FileForID
// share this mapping. Paths never leave the module.
type Catalog struct {
	root    string
	settle  time.Duration
	now     func() time.Time
	store   *store
	mu      sync.RWMutex
	files   []File
	byID    map[string]File
	pending map[string]pendingFile
	subs    map[chan struct{}]struct{}
}

// Open walks the watched folder once and admits files already there. An empty
// or missing root is a valid empty catalog. dbPath may be empty to skip
// SQLite persistence (tests). New files then wait until their size is stable.
func Open(root, dbPath string) (*Catalog, error) {
	return open(root, dbPath, 2*time.Second, time.Now)
}

// OpenImmediate admits every current file on each scan. Used by tests.
func OpenImmediate(root, dbPath string) (*Catalog, error) {
	return open(root, dbPath, 0, time.Now)
}

func open(root, dbPath string, settle time.Duration, now func() time.Time) (*Catalog, error) {
	root = strings.TrimSpace(root)
	abs := ""
	if root != "" {
		var err error
		abs, err = filepath.Abs(root)
		if err != nil {
			return nil, err
		}
	}
	c := &Catalog{
		root:    abs,
		settle:  settle,
		now:     now,
		byID:    map[string]File{},
		pending: map[string]pendingFile{},
		subs:    map[chan struct{}]struct{}{},
	}
	if strings.TrimSpace(dbPath) != "" {
		st, err := openStore(dbPath)
		if err != nil {
			return nil, err
		}
		c.store = st
	}
	if err := c.scan(true); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

// Close releases the SQLite handle.
func (c *Catalog) Close() error {
	if c == nil || c.store == nil {
		return nil
	}
	return c.store.close()
}

// Items returns the current snapshot without Mac paths.
func (c *Catalog) Items() []Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := make([]Item, len(c.files))
	for i, file := range c.files {
		items[i] = file.Item
	}
	return items
}

// FileByID returns the Mac file that owns a Nestplay id.
func (c *Catalog) FileByID(id string) (File, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	file, ok := c.byID[id]
	if !ok {
		return File{}, ErrNotFound
	}
	return file, nil
}

func (c *Catalog) broadcastLocked() {
	for ch := range c.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// Subscribe receives a coalesced signal when the snapshot changes. Call the
// returned function to unsubscribe.
func (c *Catalog) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	c.mu.Lock()
	c.subs[ch] = struct{}{}
	c.mu.Unlock()
	return ch, func() {
		c.mu.Lock()
		delete(c.subs, ch)
		c.mu.Unlock()
		close(ch)
	}
}

// Scan reconciles the snapshot with the watched folder. New or growing files
// wait until their size is stable for the settle window.
func (c *Catalog) Scan() error {
	return c.scan(false)
}

func (c *Catalog) scan(admitAll bool) error {
	disk, err := walkFiles(c.root)
	if err != nil {
		return err
	}
	now := c.now()
	next := make([]File, 0, len(disk))
	nextByID := make(map[string]File, len(disk))
	nextPending := map[string]pendingFile{}
	changed := false

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, file := range disk {
		prev, had := c.byID[file.Item.ID]
		pend, waiting := c.pending[file.Item.ID]
		switch {
		case admitAll || c.settle <= 0:
			if had {
				file.added = prev.added
			}
			next = append(next, file)
			nextByID[file.Item.ID] = file
			if !had || prev.size != file.size || prev.mtime != file.mtime {
				changed = true
			}
		case had && prev.size == file.size && prev.mtime == file.mtime:
			next = append(next, prev)
			nextByID[file.Item.ID] = prev
		case waiting && pend.size == file.size && now.Sub(pend.stableSince) >= c.settle:
			if had {
				file.added = prev.added
			}
			next = append(next, file)
			nextByID[file.Item.ID] = file
			changed = true
		default:
			if waiting && pend.size == file.size {
				nextPending[file.Item.ID] = pend
			} else {
				nextPending[file.Item.ID] = pendingFile{file: file, size: file.size, stableSince: now}
			}
			if had {
				next = append(next, prev)
				nextByID[file.Item.ID] = prev
			}
		}
	}
	if !changed {
		if len(c.byID) != len(nextByID) {
			changed = true
		} else {
			for id := range c.byID {
				if _, ok := nextByID[id]; !ok {
					changed = true
					break
				}
			}
		}
	}
	sort.SliceStable(next, func(i, j int) bool {
		if next[i].added != next[j].added {
			return next[i].added > next[j].added
		}
		if next[i].Item.Title != next[j].Item.Title {
			return next[i].Item.Title < next[j].Item.Title
		}
		return next[i].Item.Filename < next[j].Item.Filename
	})
	c.files = next
	c.byID = nextByID
	c.pending = nextPending
	if c.store != nil {
		if err := c.store.replaceAll(next); err != nil {
			return err
		}
	}
	if changed {
		c.broadcastLocked()
	}
	return nil
}

// List walks a watched folder. An empty or missing root is a valid empty
// catalog so the health check can stay up before the user adds movies.
func List(root string) ([]Item, error) {
	c, err := open(root, "", 0, time.Now)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return c.Items(), nil
}

// Files returns catalog items with absolute Mac paths. Callers must not
// put Path on the wire.
func Files(root string) ([]File, error) {
	c, err := open(root, "", 0, time.Now)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]File, len(c.files))
	copy(out, c.files)
	return out, nil
}

// FileForID returns the Mac file that owns a Nestplay id.
func FileForID(root, id string) (File, error) {
	c, err := open(root, "", 0, time.Now)
	if err != nil {
		return File{}, err
	}
	defer c.Close()
	return c.FileByID(id)
}

func walkFiles(root string) ([]File, error) {
	if root == "" {
		return []File{}, nil
	}
	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		return []File{}, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []File{}, nil
	}

	var files []File
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if skipName(d.Name()) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if depth(rel) > maxDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if _, ok := videoExt[ext]; !ok {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".part") {
			return nil
		}
		stat, err := d.Info()
		if err != nil {
			return err
		}
		title := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		slashRel := filepath.ToSlash(rel)
		files = append(files, File{
			Item: Item{
				ID:         stableID(rel),
				Title:      title,
				Filename:   d.Name(),
				Collection: collectionOf(slashRel),
			},
			Path:  path,
			rel:   slashRel,
			size:  stat.Size(),
			mtime: stat.ModTime().UnixNano(),
			added: addedAt(stat),
		})
		return nil
	})
	return files, err
}

func addedAt(info fs.FileInfo) int64 {
	if info == nil {
		return 0
	}
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		changed := time.Unix(st.Ctimespec.Sec, st.Ctimespec.Nsec)
		if !changed.IsZero() {
			return changed.UnixNano()
		}
	}
	return info.ModTime().UnixNano()
}

func skipName(name string) bool {
	return strings.HasPrefix(name, ".")
}

func collectionOf(slashRel string) string {
	i := strings.LastIndex(slashRel, "/")
	if i <= 0 {
		return ""
	}
	return slashRel[:i]
}

func depth(rel string) int {
	cleaned := filepath.ToSlash(rel)
	if cleaned == "." || cleaned == "" {
		return 0
	}
	return strings.Count(cleaned, "/") + 1
}

func stableID(rel string) string {
	sum := sha256.Sum256([]byte(filepath.ToSlash(rel)))
	return hex.EncodeToString(sum[:16])
}

// DefaultDBPath is the Mac-side SQLite file for catalog identity.
func DefaultDBPath() string {
	root, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(root) == "" {
		return ""
	}
	return filepath.Join(root, "nestplay", "library.sqlite")
}

// Watch watches the watched folder until ctx is done.
func (c *Catalog) Watch(ctx context.Context) {
	watch(ctx, c)
}
