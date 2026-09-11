package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

// ErrNotFound means no watched-folder file owns that Nestplay id.
var ErrNotFound = errors.New("catalog item not found")

// Item is a Nestplay-owned catalog record. It is not a TMDB or filesystem
// schema: the Apple TV only sees these fields.
type Item struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Filename string `json:"filename"`
}

// File is a Mac-side catalog record. Path must never be encoded on the wire.
type File struct {
	Item Item
	Path string
}

// List walks a watched folder. An empty or missing root is a valid empty
// catalog so the health check can stay up before the user adds movies.
func List(root string) ([]Item, error) {
	files, err := Files(root)
	if err != nil {
		return nil, err
	}
	items := make([]Item, len(files))
	for i, file := range files {
		items[i] = file.Item
	}
	return items, nil
}

// Files returns catalog items with absolute Mac paths. Callers must not
// put Path on the wire.
func Files(root string) ([]File, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return []File{}, nil
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
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
		files = append(files, File{
			Item: Item{
				ID:       stableID(rel),
				Title:    title,
				Filename: d.Name(),
			},
			Path: path,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].Item.Title == files[j].Item.Title {
			return files[i].Item.Filename < files[j].Item.Filename
		}
		return files[i].Item.Title < files[j].Item.Title
	})
	return files, nil
}

// FileForID returns the Mac file that owns a Nestplay id.
func FileForID(root, id string) (File, error) {
	files, err := Files(root)
	if err != nil {
		return File{}, err
	}
	for _, file := range files {
		if file.Item.ID == id {
			return file, nil
		}
	}
	return File{}, ErrNotFound
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
