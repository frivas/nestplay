package downloads

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"nestplay.local/server/internal/storage"
	"nestplay.local/server/internal/youtube"
)

const (
	StateQueued       = "queued"
	StateDownloading  = "downloading"
	StateConverting   = "converting"
	StateWaitingSpace = "waiting_space"
	StateTooLow       = "too_low"
	StateFailed       = "failed"
	StateComplete     = "complete"
)

// Item is a download-queue record. Paths never appear here.
type Item struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	State   string `json:"state"`
	Message string `json:"message,omitempty"`
}

// Fetcher downloads a YouTube id into destDir and returns the source file.
type Fetcher interface {
	Fetch(ctx context.Context, id, destDir string) (src string, err error)
}

// Converter writes a Living Room–playable MP4 at dest from src.
type Converter interface {
	MakePlayable(ctx context.Context, src, dest string) error
}

// Disk reports free bytes on the watched-folder volume.
type Disk func(path string) (uint64, error)

// Queue holds YouTube download jobs. Completed playable files are published
// into destRoot; unplayable intermediates stay out of the catalog.
type Queue struct {
	destRoot string
	workRoot string
	fetch    Fetcher
	convert  Converter
	disk     Disk
	mu       sync.Mutex
	items    []Item
}

// Open returns an in-memory queue. destRoot is the watched folder.
func Open(destRoot, workRoot string, fetch Fetcher, convert Converter, disk Disk) *Queue {
	if disk == nil {
		disk = storage.VolumeFree
	}
	return &Queue{destRoot: destRoot, workRoot: workRoot, fetch: fetch, convert: convert, disk: disk}
}

// Enqueue records ids for download. Invalid YouTube ids are ignored.
// Already-queued ids are left as-is so Home stays usable while adding more.
func (q *Queue) Enqueue(items []Item) []Item {
	if q == nil {
		return nil
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	seen := map[string]int{}
	for i, it := range q.items {
		seen[it.ID] = i
	}
	for _, it := range items {
		if !youtube.VideoID(it.ID) {
			continue
		}
		if _, ok := seen[it.ID]; ok {
			continue
		}
		if it.Title == "" {
			it.Title = it.ID
		}
		it.State = StateQueued
		it.Message = ""
		q.items = append(q.items, it)
		seen[it.ID] = len(q.items) - 1
	}
	out := make([]Item, len(q.items))
	copy(out, q.items)
	return out
}

// List returns a copy of queue items. No filesystem paths.
func (q *Queue) List() []Item {
	if q == nil {
		return []Item{}
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]Item, len(q.items))
	copy(out, q.items)
	return out
}

func (q *Queue) set(id, state, message string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i := range q.items {
		if q.items[i].ID == id {
			q.items[i].State = state
			q.items[i].Message = message
			log.Printf("queue %s %s", id, state)
			return
		}
	}
}

func (q *Queue) nextQueued() (Item, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, it := range q.items {
		if it.State == StateQueued || it.State == StateWaitingSpace {
			return it, true
		}
	}
	return Item{}, false
}

// ProcessOne runs a single queued job. Tests call this; Run loops it.
func (q *Queue) ProcessOne(ctx context.Context) error {
	if q == nil || q.fetch == nil || q.convert == nil {
		return errors.New("queue not configured")
	}
	item, ok := q.nextQueued()
	if !ok {
		return nil
	}
	if q.destRoot == "" {
		q.set(item.ID, StateFailed, "No watched folder.")
		return nil
	}
	free, err := q.disk(q.destRoot)
	if err != nil || storage.Admit(free) != nil {
		q.set(item.ID, StateWaitingSpace, "Waiting for disk space.")
		return storage.ErrRefused
	}
	if err := os.MkdirAll(q.workRoot, 0o700); err != nil {
		log.Printf("queue %s work: %v", item.ID, err)
		q.set(item.ID, StateFailed, "Could not start download.")
		return err
	}
	work, err := os.MkdirTemp(q.workRoot, "yt-")
	if err != nil {
		log.Printf("queue %s work: %v", item.ID, err)
		q.set(item.ID, StateFailed, "Could not start download.")
		return err
	}
	defer os.RemoveAll(work)

	q.set(item.ID, StateDownloading, "")
	src, err := q.fetch.Fetch(ctx, item.ID, work)
	if errors.Is(err, ErrBelow720) {
		q.set(item.ID, StateTooLow, "YouTube has no 720p or higher.")
		return err
	}
	if err != nil {
		if errors.Is(err, storage.ErrRefused) || errors.Is(ctx.Err(), context.Canceled) {
			q.set(item.ID, StateWaitingSpace, "Waiting for disk space.")
			return err
		}
		log.Printf("queue %s fetch: %v", item.ID, err)
		q.set(item.ID, StateFailed, "Download failed.")
		return err
	}
	free, err = q.disk(q.destRoot)
	if err != nil || storage.Admit(free) != nil {
		q.set(item.ID, StateWaitingSpace, "Waiting for disk space.")
		return storage.ErrRefused
	}
	q.set(item.ID, StateConverting, "")
	name := sanitizeName(item.Title) + "-" + item.ID + ".mp4"
	dest := filepath.Join(q.destRoot, name)
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		q.set(item.ID, StateFailed, "Could not publish the file.")
		return err
	}
	tmp := filepath.Join(work, "playable.mp4")
	if err := q.convert.MakePlayable(ctx, src, tmp); err != nil {
		log.Printf("queue %s convert: %v", item.ID, err)
		q.set(item.ID, StateFailed, "Could not make this video playable.")
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		q.set(item.ID, StateFailed, "Could not publish the file.")
		return err
	}
	q.set(item.ID, StateComplete, "")
	return nil
}

// Run processes the queue until ctx is done.
func (q *Queue) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = q.ProcessOne(ctx)
		}
	}
}

func sanitizeName(title string) string {
	if title == "" {
		return "video"
	}
	var b []rune
	for _, r := range title {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', '\n', '\r', 0:
			b = append(b, '-')
		default:
			b = append(b, r)
		}
	}
	name := string(b)
	if len(name) > 80 {
		name = name[:80]
	}
	return name
}

// DecodeEnqueue reads POST /queue JSON.
func DecodeEnqueue(body []byte) ([]Item, error) {
	var payload struct {
		Items []Item `json:"items"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return payload.Items, nil
}
