package storage

import (
	"errors"
	"syscall"
)

// Reserve is the free-space floor. New downloads and conversion writes stop
// here so playback of completed local-library files can continue.
const Reserve = 20 * 1024 * 1024 * 1024

// ErrRefused means a writer must not start or must abort.
var ErrRefused = errors.New("storage reserve refused")

// Admit reports whether a writer may run given free bytes on the volume.
func Admit(free uint64) error {
	if free <= Reserve {
		return ErrRefused
	}
	return nil
}

// VolumeFree returns available bytes on the volume that contains path.
func VolumeFree(path string) (uint64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, err
	}
	return uint64(st.Bavail) * uint64(st.Bsize), nil
}
