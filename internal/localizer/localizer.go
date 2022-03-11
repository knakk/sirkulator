package localizer

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	_ "github.com/knakk/sirkulator/internal/translations"
)

type Localizer struct {
	Lang    language.Tag
	printer *message.Printer
}

var locales = []Localizer{
	{
		Lang:    language.English,
		printer: message.NewPrinter(language.English),
	},
	{
		Lang:    language.Norwegian,
		printer: message.NewPrinter(language.Norwegian),
	},
}

var (
	// Matcher will return the first supported language from the give language tags.
	Matcher = language.NewMatcher([]language.Tag{language.English, language.Norwegian})

	defaultLocale = 0 // en
)

// SetDefaltLocale sets the defaul locale. It will panic if
// the given tag is not among locales.
// It is not safe for concurrent use, and should be set once
// at program startup.
func SetDefaultLocale(lang language.Tag) {
	found := false
	var tags []language.Tag
	for i, locale := range locales {
		if lang == locale.Lang {
			defaultLocale = i
			found = true
		}
	}
	if !found {
		panic("localizer: unsupported language:" + lang.String())
	}
	tags = append(tags, locales[defaultLocale].Lang)
	for i, locale := range locales {
		if i != defaultLocale {
			tags = append(tags, locale.Lang)
		}
	}
	Matcher = language.NewMatcher(tags)
}

func Get(lang language.Tag) Localizer {
	for _, locale := range locales {
		if lang.Parent() == locale.Lang {
			return locale
		}
	}

	return locales[defaultLocale]
}

// GetFromAccept lang returns a Localizer parsed from the given Accept-language header string.
func GetFromAcceptLang(lang string) Localizer {
	t, _, _ := language.ParseAcceptLanguage(lang)
	// We ignore the error: the default language will be selected for t == nil.
	tag, _, _ := Matcher.Match(t...)
	return Get(tag)
}

func (l Localizer) Translate(key message.Reference, args ...interface{}) string {
	return l.printer.Sprintf(key, args...)
}
