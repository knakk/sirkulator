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
	First     string
	FirstID   string
	Last      string
	LastID    string
	From      string
	FromID    string
	To        string
	ToID      string
	Target    string
	PrevLabel string
	NextLabel string
	HasMore   bool
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

	if t.From != "" || t.To != "" && t.HasMore {
		// We're not on first page
		writeAttr(w, "class", "clickable")
		io.WriteString(w, `><a hx-swap="outerHTML" hx-target="#`)
		io.WriteString(w, t.ID)
		io.WriteString(w, `"`)
		writeAttr(w, "hx-get", fmt.Sprintf("%slimit=%d&to=%s&to_id=%s&sort_by=%s&sort_dir=%s", t.Target, t.Limit, t.First, t.FirstID, t.SortBy, t.SortDir))
		io.WriteString(w, ">ᐊ ")
		io.WriteString(w, t.PrevLabel)
		io.WriteString(w, "</a>")
	} else {
		io.WriteString(w, ">")
	}
	io.WriteString(w, "</td><td")

	if t.HasMore || t.To != "" {
		writeAttr(w, "class", "clickable")
		io.WriteString(w, `><a hx-swap="outerHTML" hx-target="#`)
		io.WriteString(w, t.ID)
		io.WriteString(w, `"`)
		writeAttr(w, "hx-get", fmt.Sprintf("%slimit=%d&from=%s&from_id=%s&sort_by=%s&sort_dir=%s", t.Target, t.Limit, t.Last, t.LastID, t.SortBy, t.SortDir))
		io.WriteString(w, ">")
		io.WriteString(w, t.NextLabel)
		io.WriteString(w, " ᐅ</a>")
	} else {
		io.WriteString(w, ">")
	}

	io.WriteString(w, "</td></tr></tfoot>")
}
