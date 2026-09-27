package downloads

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"nestplay.local/server/internal/catalog"
	"nestplay.local/server/internal/storage"
)

// ConvertHidden turns one hidden catalog file into a Living Room MP4.
// On failure the original is left in place. Temps stay under workRoot.
func ConvertHidden(ctx context.Context, cat *catalog.Catalog, convert Converter, disk Disk, workRoot string) error {
	if cat == nil || convert == nil {
		return nil
	}
	if disk == nil {
		disk = storage.VolumeFree
	}
	hidden := cat.Hidden()
	if len(hidden) == 0 {
		return nil
	}
	file := hidden[0]
	free, err := disk(file.Path)
	if err != nil || storage.Admit(free) != nil {
		return storage.ErrRefused
	}
	if workRoot == "" {
		workRoot = filepath.Join(os.TempDir(), "nestplay-work")
	}
	if err := os.MkdirAll(workRoot, 0o700); err != nil {
		return err
	}
	work, err := os.MkdirTemp(workRoot, "conv-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	log.Printf("convert %s", file.Path)
	base := strings.TrimSuffix(file.Path, filepath.Ext(file.Path))
	dest := base + ".mp4"
	tmp := filepath.Join(work, "playable.mp4")
	if err := convert.MakePlayable(ctx, file.Path, tmp); err != nil {
		log.Printf("convert %s: %v", file.Path, err)
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		return err
	}
	if dest != file.Path {
		_ = os.Remove(file.Path)
	}
	return nil
}
