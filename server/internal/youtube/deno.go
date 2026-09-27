package youtube

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	_ "embed"
)

//go:embed youtubejs.ts
var youtubeJS []byte

//go:embed youtube_items.ts
var youtubeItems []byte

type denoAdapter struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	stdin   *bufio.Writer
	stdout  *bufio.Scanner
	cookie  string
	started bool
}

func New(cookieFile string) Adapter {
	cookieFile = filepath.Clean(cookieFile)
	data, err := os.ReadFile(cookieFile) // #nosec G304 -- cookie path is a local Nestplay config file, never request input
	if err != nil {
		return Disabled()
	}
	header := cookieHeader(string(data))
	if header == "" {
		return Disabled()
	}
	headerPath, err := persistHeader(header)
	if err != nil {
		return Disabled()
	}
	return &denoAdapter{cookie: headerPath}
}

func persistHeader(header string) (string, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "nestplay")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "youtube-cookies.header")
	if err := os.WriteFile(path, []byte(header), 0o600); err != nil { // #nosec G703 -- path is a fixed file under the Nestplay cache directory
		return "", err
	}
	return path, nil
}

func (a *denoAdapter) Status(ctx context.Context) Page {
	return a.roundTrip(ctx, map[string]string{"op": "status"})
}

func (a *denoAdapter) Home(ctx context.Context, cursor string, _ bool) Page {
	req := map[string]string{"op": "home"}
	if cursor != "" {
		req["cursor"] = cursor
	}
	return a.roundTrip(ctx, req)
}

func (a *denoAdapter) Search(ctx context.Context, query, cursor string) Page {
	req := map[string]string{"op": "search", "query": query}
	if cursor != "" {
		req["cursor"] = cursor
	}
	return a.roundTrip(ctx, req)
}

func (a *denoAdapter) roundTrip(ctx context.Context, req map[string]string) Page {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.ensure(); err != nil {
		return Page{State: Unavailable, Message: "YouTube Home could not be loaded."}
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return Page{State: Unavailable, Message: "YouTube Home could not be loaded."}
	}
	if _, err := a.stdin.Write(append(payload, '\n')); err != nil {
		a.reset()
		return Page{State: Unavailable, Message: "YouTube Home could not be loaded."}
	}
	if err := a.stdin.Flush(); err != nil {
		a.reset()
		return Page{State: Unavailable, Message: "YouTube Home could not be loaded."}
	}
	type reply struct {
		OK      bool   `json:"ok"`
		State   State  `json:"state"`
		Message string `json:"message"`
		Cursor  string `json:"cursor"`
		Items   []Item `json:"items"`
	}
	done := make(chan reply, 1)
	go func() {
		if !a.stdout.Scan() {
			done <- reply{State: Unavailable, Message: "YouTube Home could not be loaded."}
			return
		}
		var out reply
		if err := json.Unmarshal(a.stdout.Bytes(), &out); err != nil {
			done <- reply{State: Unavailable, Message: "YouTube Home could not be loaded."}
			return
		}
		if out.State == "" {
			out.State = Connected
		}
		done <- out
	}()
	select {
	case <-ctx.Done():
		a.reset()
		return Page{State: Unavailable, Message: "YouTube Home could not be loaded."}
	case out := <-done:
		if out.State == "" {
			out.State = Reconnect
		}
		return Page{State: out.State, Message: out.Message, Items: out.Items, Cursor: out.Cursor}
	}
}

func (a *denoAdapter) ensure() error {
	if a.started && a.cmd != nil && a.cmd.Process != nil {
		return nil
	}
	deno, err := exec.LookPath("deno")
	if err != nil {
		return err
	}
	script, err := writeScript()
	if err != nil {
		return err
	}
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, deno, // #nosec G204 -- deno path comes from LookPath; script is the embedded adapter
		"run",
		"--quiet",
		"--no-prompt",
		"--allow-net",
		"--allow-env=NESTPLAY_YOUTUBE_COOKIES_FILE",
		"--allow-read="+a.cookie,
		script,
	)
	cmd.Env = append(os.Environ(), "NESTPLAY_YOUTUBE_COOKIES_FILE="+a.cookie)
	cmd.Stderr = &bytes.Buffer{}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	a.cmd = cmd
	a.stdin = bufio.NewWriter(stdin)
	a.stdout = bufio.NewScanner(stdout)
	a.stdout.Buffer(make([]byte, 0, 64*1024), 4<<20)
	a.started = true
	return nil
}

func (a *denoAdapter) reset() {
	if a.cmd != nil && a.cmd.Process != nil {
		_ = a.cmd.Process.Kill()
		_, _ = a.cmd.Process.Wait()
	}
	a.cmd = nil
	a.stdin = nil
	a.stdout = nil
	a.started = false
}

func writeScript() (string, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "nestplay")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "youtube_items.ts"), youtubeItems, 0o600); err != nil { // #nosec G306 -- adapter script is a fixed file under the Nestplay cache directory
		return "", err
	}
	path := filepath.Join(dir, "youtubejs.ts")
	if err := os.WriteFile(path, youtubeJS, 0o600); err != nil { // #nosec G306 -- adapter script is a fixed file under the Nestplay cache directory
		return "", err
	}
	return path, nil
}

func DefaultCookieFile() string {
	root, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(root, "nestplay", "youtube-cookies.txt")
}

func VideoID(id string) bool {
	if len(id) != 11 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_' {
			continue
		}
		return false
	}
	return true
}
