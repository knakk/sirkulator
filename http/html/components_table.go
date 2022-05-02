package html

import (
	"context"
	"fmt"
	"io"
)

type TablePaginated struct {
	ID        string
	Class     string
	SortBy    string
	SortDir   string
	Target    string
	PrevLabel string
	NextLabel string
	HasMore   bool
	Offset    int
	Limit     int
	Yield     func()
}

func (t *TablePaginated) Render(ctx context.Context, w io.Writer) {
	io.WriteString(w, "<table")
	writeAttr(w, "id", t.ID)
	writeAttr(w, "class", t.Class)
	io.WriteString(w, ">")

	// TODO put theese hidden inputs in tfoot
	if t.SortBy != "" {
		io.WriteString(w, `<input type="hidden" name="sort_by"`)
		writeAttr(w, "value", t.SortBy)
		io.WriteString(w, `><input type="hidden" name="sort_dir"`)
		writeAttr(w, "value", t.SortDir)
		io.WriteString(w, ">")
	}

	if t.Yield != nil {
		t.Yield()
	}

	io.WriteString(w, `<tfoot class="pagination"><tr><td`)

	if t.Offset > 0 {
		// We're not on first page
		writeAttr(w, "class", "clickable")
		io.WriteString(w, `><a hx-swap="outerHTML" hx-target="#`)
		io.WriteString(w, t.ID)
		io.WriteString(w, `"`)
		writeAttr(w, "hx-get", fmt.Sprintf("%slimit=%d&offset=%d&sort_by=%s&sort_dir=%s", t.Target, t.Limit, t.Offset-t.Limit, t.SortBy, t.SortDir))
		io.WriteString(w, ">ᐊ ")
		io.WriteString(w, t.PrevLabel)
		io.WriteString(w, "</a>")
	} else {
		io.WriteString(w, ">")
	}
	io.WriteString(w, "</td><td")

	if t.HasMore {
		writeAttr(w, "class", "clickable")
		io.WriteString(w, `><a hx-swap="outerHTML" hx-target="#`)
		io.WriteString(w, t.ID)
		io.WriteString(w, `"`)
		writeAttr(w, "hx-get", fmt.Sprintf("%slimit=%d&offset=%d&sort_by=%s&sort_dir=%s", t.Target, t.Limit, t.Offset+t.Limit, t.SortBy, t.SortDir))
		io.WriteString(w, ">")
		io.WriteString(w, t.NextLabel)
		io.WriteString(w, " ᐅ</a>")
	} else {
		io.WriteString(w, ">")
	}

	io.WriteString(w, "</td></tr></tfoot>")
}
