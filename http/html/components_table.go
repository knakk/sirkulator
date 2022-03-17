package html

import (
	"context"
	"fmt"
	"io"
)

type TablePaginated struct {
	ID        string
	Class     string
	First     string
	Last      string
	From      string
	To        string
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

	if t.Yield != nil {
		t.Yield()
	}

	io.WriteString(w, `<tfoot class="pagination"><tr><td`)

	if t.From != "" || t.To != "" && t.HasMore {
		// We're not on first page
		writeAttr(w, "class", "clickable")
		io.WriteString(w, `><a hx-swap="outerHTML" hx-target="#parts-of"`)
		writeAttr(w, "hx-get", fmt.Sprintf("%s?limit=%d&to=%s", t.Target, t.Limit, t.First))
		io.WriteString(w, ">ᐊ ")
		io.WriteString(w, t.PrevLabel)
		io.WriteString(w, "</a>")
	} else {
		io.WriteString(w, ">")
	}
	io.WriteString(w, "</td><td")

	if t.HasMore || t.To != "" {
		writeAttr(w, "class", "clickable")
		io.WriteString(w, `><a hx-swap="outerHTML" hx-target="#parts-of"`)
		writeAttr(w, "hx-get", fmt.Sprintf("%s?limit=%d&from=%s", t.Target, t.Limit, t.Last))
		io.WriteString(w, ">")
		io.WriteString(w, t.NextLabel)
		io.WriteString(w, " ᐅ</a>")
	} else {
		io.WriteString(w, ">")
	}

	io.WriteString(w, "</td></tr></tfoot>")
}
