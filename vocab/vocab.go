package vocab

import (
	"errors"

	"golang.org/x/text/language"
)

// Term is a vocabulary term.
type Term interface {
	// Code is the coded value of a Term.
	Code() string

	// URI is the canonical representation of the Term.
	// The URI does not need to resolve.
	URI() string

	// Label should return a localized string representation of a Term.
	Label(language.Tag) string

	//Alias(language.Tag) []string
}

var (
	ErrDeprecated = errors.New("vocab: deprecated")
	ErrUnknown    = errors.New("vocab: unknown")
)
