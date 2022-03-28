package html

import (
	"context"
	"fmt"
	"html"
	"io"
	"strconv"
	"strings"
)

func writeAttr(w io.Writer, key, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(w, ` %s=%q`, key, html.EscapeString(value))
}

type InputString struct {
	ID            string
	Label         string
	Size          string // int
	Value         string
	Validation    string
	ValidationMsg string
	Required      bool
}

func (i *InputString) Render(ctx context.Context, w io.Writer) {
	if i.ValidationMsg == "" {
		i.ValidationMsg = "&nbsp;"
	} else {
		i.ValidationMsg = "▲ " + i.ValidationMsg
	}
	io.WriteString(w, `<div class="field"><input type="text" autocomplete="off"`)
	writeAttr(w, "id", i.ID)
	writeAttr(w, "name", i.ID)
	writeAttr(w, "size", i.Size)
	writeAttr(w, "maxlength", i.Size)
	writeAttr(w, "value", i.Value)
	if i.Required {
		io.WriteString(w, ` required`)
	}
	if i.Validation != "" {
		fmt.Fprintf(w, ` %s="%s"`, "pattern", i.Validation)
	}
	io.WriteString(w, `><label`)
	writeAttr(w, "for", i.ID)
	io.WriteString(w, `>`)
	io.WriteString(w, i.Label)
	io.WriteString(w, `</label><p class="validation"><small>`)
	io.WriteString(w, i.ValidationMsg)
	io.WriteString(w, `</small></p></div>`)
}

type InputText struct {
	ID       string
	Label    string
	Rows     int
	Value    string
	InfoMsg  string
	Required bool
}

func (i *InputText) Render(ctx context.Context, w io.Writer) {
	if i.InfoMsg == "" {
		i.InfoMsg = "&nbsp;"
	}
	if i.Rows < 2 {
		i.Rows = 2
	}
	io.WriteString(w, `<div class="field"><textarea`)
	writeAttr(w, "id", i.ID)
	writeAttr(w, "name", i.ID)
	writeAttr(w, "rows", strconv.Itoa(i.Rows))
	if i.Required {
		io.WriteString(w, ` required`)
	}
	io.WriteString(w, `>`)
	io.WriteString(w, html.EscapeString(i.Value))
	io.WriteString(w, `</textarea><label`)
	writeAttr(w, "for", i.ID)
	io.WriteString(w, `>`)
	io.WriteString(w, i.Label)
	io.WriteString(w, `</label><p class="info"><small>`)
	io.WriteString(w, i.InfoMsg)
	io.WriteString(w, `</small></p></div>`)
}

type InputBool struct {
	ID      string
	Label   string
	Rows    int
	Value   bool
	InfoMsg string
}

func (i *InputBool) Render(ctx context.Context, w io.Writer) {
	if i.InfoMsg == "" {
		i.InfoMsg = "&nbsp;"
	}
	io.WriteString(w, `<div class="field"><input type="checkbox"`)
	writeAttr(w, "id", i.ID)
	writeAttr(w, "name", i.ID)
	if i.Value {
		io.WriteString(w, ` checked`)
	}
	io.WriteString(w, `><label`)
	writeAttr(w, "for", i.ID)
	io.WriteString(w, `>`)
	io.WriteString(w, i.Label)
	io.WriteString(w, `</label><p class="info"><small>`)
	io.WriteString(w, i.InfoMsg)
	io.WriteString(w, `</small></p></div>`)
}

type InputRadio struct {
	ID       string
	Label    string
	Options  [][2]string
	Value    string
	InfoMsg  string
	Required bool
}

func (i *InputRadio) Render(ctx context.Context, w io.Writer) {
	if i.InfoMsg == "" {
		i.InfoMsg = "&nbsp;"
	}
	io.WriteString(w, `<div class="radiofield"><div class="label">`)
	io.WriteString(w, html.EscapeString(i.Label))
	io.WriteString(w, `</div><div class="radiooptions">`)
	for n, opt := range i.Options {
		io.WriteString(w, `<input type="radio"`)
		fmt.Fprintf(w, ` id="%s_%d"`, i.ID, n)
		writeAttr(w, "name", i.ID)
		writeAttr(w, "value", opt[0])
		if i.Value == opt[0] {
			io.WriteString(w, ` checked`)
		}
		if i.Required {
			io.WriteString(w, ` required`)
		}
		io.WriteString(w, `><label`)
		fmt.Fprintf(w, ` for="%s_%d"`, i.ID, n)
		io.WriteString(w, `>`)
		io.WriteString(w, opt[1])
		io.WriteString(w, `</label>`)
	}
	io.WriteString(w, `<p class="info"><small>`)
	io.WriteString(w, i.InfoMsg)
	io.WriteString(w, `</small></p></div></div>`)
}

type SearchSelect struct {
	ID        string
	Label     string
	Options   [][2]string
	URIPrefix string
	Value     string
	Values    []string
	InfoMsg   string
	Multiple  bool
	Required  bool
}

func label(opts [][2]string, val string) (string, bool) {
	for _, opt := range opts {
		if opt[0] == val {
			return opt[1], true
		}
	}
	return "", false
}

func sliceContains(slice []string, v string) bool {
	for _, s := range slice {
		if s == v {
			return true
		}
	}
	return false
}

func (i *SearchSelect) Render(ctx context.Context, w io.Writer) {
	if i.InfoMsg == "" {
		i.InfoMsg = "&nbsp;"
	}
	io.WriteString(w, `<div class="field"><input class="search-select`)
	if !i.Multiple && i.Value != "" {
		io.WriteString(w, ` single-value" disabled`)
		i.Values = []string{i.Value}
	} else {
		io.WriteString(w, `"`)
	}
	io.WriteString(w, ` type="search" autocomplete="off"`)
	writeAttr(w, "id", i.ID)
	writeAttr(w, "list", i.ID+"_list")
	io.WriteString(w, `><datalist`)
	writeAttr(w, "id", i.ID+"_list")
	io.WriteString(w, `>`)
	for _, opt := range i.Options {
		if sliceContains(i.Values, i.URIPrefix+opt[0]) {
			// exclude already selected values from datalist
			continue
		}
		io.WriteString(w, `<option`)
		writeAttr(w, "value", opt[0])
		io.WriteString(w, `>`)
		io.WriteString(w, html.EscapeString(opt[1]))
		io.WriteString(w, `</option>`)
	}
	io.WriteString(w, `</datalist><label`)
	writeAttr(w, "for", i.ID)
	io.WriteString(w, `>`)
	io.WriteString(w, i.Label)
	io.WriteString(w, `</label><ul>`)
	for _, v := range i.Values {
		v = strings.TrimPrefix(v, i.URIPrefix)
		if label, ok := label(i.Options, v); ok {
			io.WriteString(w, `<li><div class="selected-term">`)
			io.WriteString(w, label)
			io.WriteString(w, `</div><input type="hidden"`)
			writeAttr(w, "name", i.ID)
			writeAttr(w, "value", v)
			io.WriteString(w, `>`)
			io.WriteString(w, `<button type="button" class="unselect-term"><span>✕</span></button></li>`)
		}
	}
	io.WriteString(w, `</ul><p class="info"><small>`)
	io.WriteString(w, i.InfoMsg)
	io.WriteString(w, `</small></p></div>`)
}
