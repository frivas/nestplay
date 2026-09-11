package youtube

import "context"

type Fake struct {
	Now   Page
	Pages map[string]Page
}

func (f *Fake) page(key string) Page {
	if f.Pages != nil {
		if page, ok := f.Pages[key]; ok {
			return page
		}
	}
	return f.Now
}

func (f *Fake) Status(context.Context) Page { return f.page("status") }

func (f *Fake) Home(_ context.Context, cursor string) Page {
	return f.page("home:" + cursor)
}

func (f *Fake) Search(_ context.Context, query, cursor string) Page {
	return f.page("search:" + query + ":" + cursor)
}
