package catalog

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

func watch(ctx context.Context, c *Catalog) {
	if c == nil || c.root == "" {
		<-ctx.Done()
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		tickerWatch(ctx, c)
		return
	}
	defer watcher.Close()
	_ = addWatches(watcher, c.root)

	debounce := time.NewTimer(time.Hour)
	if !debounce.Stop() {
		select {
		case <-debounce.C:
		default:
		}
	}
	poll := time.NewTicker(2 * time.Second)
	defer poll.Stop()

	kick := func() {
		if !debounce.Stop() {
			select {
			case <-debounce.C:
			default:
			}
		}
		debounce.Reset(250 * time.Millisecond)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-poll.C:
			_ = addWatches(watcher, c.root)
			_ = c.Scan()
		case <-debounce.C:
			_ = addWatches(watcher, c.root)
			_ = c.Scan()
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					_ = watcher.Add(event.Name)
				}
			}
			kick()
		case _, ok := <-watcher.Errors:
			if !ok {
				return
			}
			kick()
		}
	}
}

func tickerWatch(ctx context.Context, c *Catalog) {
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			_ = c.Scan()
		}
	}
}

func addWatches(watcher *fsnotify.Watcher, root string) error {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return err
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if skipName(d.Name()) && path != root {
				return fs.SkipDir
			}
			_ = watcher.Add(path)
		}
		return nil
	})
}
