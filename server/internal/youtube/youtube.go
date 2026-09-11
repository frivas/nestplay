package youtube

import "context"

type State string

const (
	Connected   State = "connected"
	Reconnect   State = "reconnect_required"
	Unavailable State = "unavailable"
)

type Item struct {
	ID    string
	Title string
}

type Page struct {
	State   State
	Message string
	Items   []Item
	Cursor  string
}

type Adapter interface {
	Status(ctx context.Context) Page
	Home(ctx context.Context, cursor string) Page
	Search(ctx context.Context, query, cursor string) Page
}

type disabled struct{}

func Disabled() Adapter { return disabled{} }

func (disabled) Status(context.Context) Page {
	return Page{
		State:   Reconnect,
		Message: "YouTube needs a reconnect on your Mac.",
	}
}

func (d disabled) Home(ctx context.Context, _ string) Page {
	return d.Status(ctx)
}

func (d disabled) Search(ctx context.Context, _, _ string) Page {
	return d.Status(ctx)
}
