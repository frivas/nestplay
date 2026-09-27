package youtube

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const refreshHomeMessage = "Refresh to load YouTube Home."

type homeSnapshot struct {
	CookieMod int64  `json:"cookie_mod"`
	Items     []Item `json:"items"`
}

type cachedHome struct {
	inner      Adapter
	path       string
	cookiePath string
	mu         sync.Mutex
	loaded     bool
	items      []Item
	liveCursor string
	cookieMod  int64
}

func Cached(inner Adapter, cachePath, cookiePath string) Adapter {
	if inner == nil {
		inner = Disabled()
	}
	if cachePath == "" {
		return inner
	}
	return &cachedHome{inner: inner, path: cachePath, cookiePath: cookiePath}
}

func DefaultHomeCacheFile() string {
	root, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(root, "nestplay", "youtube-home.json")
}

func (c *cachedHome) Status(ctx context.Context) Page {
	return c.inner.Status(ctx)
}

func (c *cachedHome) Search(ctx context.Context, query, cursor string) Page {
	return c.inner.Search(ctx, query, cursor)
}

func (c *cachedHome) Home(ctx context.Context, cursor string, refresh bool) Page {
	if cursor != "" {
		page := c.inner.Home(ctx, cursor, false)
		if page.State == Connected {
			c.mu.Lock()
			c.loadLocked()
			c.appendLocked(page.Items)
			c.liveCursor = page.Cursor
			_ = c.writeLocked()
			c.mu.Unlock()
		}
		return page
	}
	if !refresh {
		c.mu.Lock()
		defer c.mu.Unlock()
		return c.snapshotLocked()
	}
	page := c.inner.Home(ctx, "", true)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadLocked()
	if page.State == Connected {
		c.items = append([]Item(nil), page.Items...)
		c.liveCursor = page.Cursor
		c.cookieMod = cookieMod(c.cookiePath)
		_ = c.writeLocked()
		return page
	}
	if page.State == Reconnect {
		return page
	}
	if snap := c.snapshotLocked(); len(snap.Items) > 0 {
		snap.Message = ""
		return snap
	}
	return page
}

func (c *cachedHome) snapshotLocked() Page {
	c.loadLocked()
	if c.cookiePath != "" {
		if _, err := os.Stat(c.cookiePath); err != nil {
			return Page{State: Reconnect, Message: "YouTube needs a reconnect on your Mac."}
		}
	}
	if !c.cookieFreshLocked() {
		return Page{State: Connected, Message: refreshHomeMessage}
	}
	if len(c.items) == 0 {
		return Page{State: Connected, Message: refreshHomeMessage}
	}
	return Page{State: Connected, Items: append([]Item(nil), c.items...), Cursor: c.liveCursor}
}

func (c *cachedHome) cookieFreshLocked() bool {
	if c.cookiePath == "" {
		return true
	}
	return c.cookieMod != 0 && c.cookieMod == cookieMod(c.cookiePath)
}

func (c *cachedHome) loadLocked() {
	if c.loaded {
		return
	}
	c.loaded = true
	data, err := os.ReadFile(c.path) // #nosec G304 -- cache path is a Nestplay cache file, never request input
	if err != nil {
		return
	}
	var snap homeSnapshot
	if json.Unmarshal(data, &snap) != nil {
		return
	}
	c.cookieMod = snap.CookieMod
	c.items = snap.Items
}

func (c *cachedHome) appendLocked(items []Item) {
	seen := make(map[string]struct{}, len(c.items))
	for _, item := range c.items {
		seen[item.ID] = struct{}{}
	}
	for _, item := range items {
		if item.ID == "" {
			continue
		}
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		c.items = append(c.items, item)
	}
}

func (c *cachedHome) writeLocked() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(homeSnapshot{CookieMod: c.cookieMod, Items: c.items})
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o600) // #nosec G306 -- cache path is a fixed file under the Nestplay cache directory
}

func cookieMod(path string) int64 {
	if path == "" {
		return 0
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.ModTime().UnixNano()
}
