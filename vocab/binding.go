package vocab

import (
	"sort"
	"strings"

	"github.com/knakk/sirkulator/internal/localizer"
	"golang.org/x/text/language"
)

type Binding string

const (
	BindingUnknown   Binding = "?"
	BindingHardback  Binding = "hardback"
	BindingPaperback Binding = "paperback"
	BindingSpiral    Binding = "spiral"
	BindingBoard     Binding = "board"
)

// TODO check RDA bindings:
// https://www.rdaregistry.info/termList/RDATypeOfBinding/

var allBindings = []Binding{
	BindingUnknown,
	BindingHardback,
	BindingPaperback,
	BindingSpiral,
	BindingBoard,
}

var bindingLabels = map[Binding][2]string{
	BindingUnknown:   {"Unknown", "Ukjent"},
	BindingHardback:  {"Hardback", "Innbundet"},
	BindingPaperback: {"Paperback", "Heftet"},
	BindingSpiral:    {"Spiral binding", "Spiralrygg"},
	BindingBoard:     {"Board book", "Pappbok"},
}

func ParseBinding(s string) Binding {
	switch strings.TrimSuffix(s, ".") {
	case "ib", "innbundet", "hardback":
		return BindingHardback
	case "h", "heftet", "paperback":
		return BindingPaperback
	case "spiralrygg", "spiral":
		return BindingSpiral
	case "board", "kartonert":
		return BindingBoard
	default:
		return BindingUnknown
	}
}

func (b Binding) Label(tag language.Tag) string {
	match, _, _ := localizer.Matcher.Match(tag)

	i := 0
	if match == language.Norwegian {
		i = 1
	}
	return bindingLabels[b][i]
}

func BindingOptions(lang language.Tag) (res [][2]string) {
	match, _, _ := localizer.Matcher.Match(lang)

	i := 0
	if match == language.Norwegian {
		i = 1
	}

	for _, b := range allBindings {
		res = append(res, [2]string{string(b), bindingLabels[b][i]})
	}

	// Sort by label
	sort.Slice(res, func(i, j int) bool {
		return res[i][1] < res[j][1]
	})

	return res
}
