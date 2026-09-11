package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxDepth = 6

var videoExt = map[string]struct{}{
	".mp4": {}, ".mkv": {}, ".m4v": {}, ".mov": {}, ".avi": {}, ".webm": {},
}

// Item is a Nestplay-owned catalog record. It is not a TMDB or filesystem
// schema: the Apple TV only sees these fields.
type Item struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Filename string `json:"filename"`
}

// List walks a watched folder. An empty or missing root is a valid empty
// catalog so the health check can stay up before the user adds movies.
func List(root string) ([]Item, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return []Item{}, nil
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if os.IsNotExist(err) {
		return []Item{}, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []Item{}, nil
	}

	var items []Item
	err = filepath.WalkDir(abs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
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
		title := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		items = append(items, Item{
			ID:       stableID(rel),
			Title:    title,
			Filename: d.Name(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Title == items[j].Title {
			return items[i].Filename < items[j].Filename
		}
		return items[i].Title < items[j].Title
	})
	return items, nil
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
