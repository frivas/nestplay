package downloads

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Sweep removes Nestplay scratch left after a crash or abort: the work
// directory and leftover .mp4.part files in the watched folder.
func Sweep(destRoot, workRoot string) {
	if workRoot != "" {
		_ = os.RemoveAll(workRoot)
	}
	if destRoot == "" {
		return
	}
	_ = filepath.WalkDir(destRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != destRoot && strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".mp4.part") {
			_ = os.Remove(path)
		}
		return nil
	})
}
