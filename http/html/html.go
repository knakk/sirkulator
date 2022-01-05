package html

import "golang.org/x/text/language"

type Page struct {
	Title string
	Lang  language.Tag
	Path  string
}
