package vocab

import (
	"sort"
	"strings"

	"github.com/knakk/sirkulator/internal/localizer"
	"golang.org/x/text/language"
)

// https://www.nb.no/nbvok/tg/nb/
// Vokabular for målgrupper
// Vokabular over demografiske grupper brukt til å indikere hvem noe er ment eller laget for.

type Audience string

const (
	TG1000 Audience = "TG1000"
	TG1001 Audience = "TG1001"
	TG1002 Audience = "TG1002"
	TG1003 Audience = "TG1003"
	TG1004 Audience = "TG1004"
	TG1005 Audience = "TG1005"
	TG1006 Audience = "TG1006"
	TG1007 Audience = "TG1007"
	TG1008 Audience = "TG1008"
	TG1009 Audience = "TG1009"
	TG1010 Audience = "TG1010"
	TG1011 Audience = "TG1011"
	TG1012 Audience = "TG1012"
	TG1013 Audience = "TG1013"
	TG1014 Audience = "TG1014"
	TG1015 Audience = "TG1015"
	TG1016 Audience = "TG1016"
	TG1017 Audience = "TG1017"
)

var audienceLabels = map[Audience][2]string{
	// Aldersgrupper:
	TG1000: {"0-2 years", "0-2 år"},
	TG1001: {"3-5 years", "3-5 år"},
	TG1002: {"6-8 years", "6-8 år"},
	TG1003: {"9-10 years", "9-10 år"},
	TG1004: {"11-12 years", "11-12 år"},
	TG1005: {"13-15 years", "13-15 år"},
	TG1015: {"16-17 years", "16-17 år"},
	TG1016: {"18 years and over", "18 år og oppover"},
	// Grupper med spesielle behov:
	TG1006: {"Easy reader", "Lettlest"},
	TG1007: {"Simple content", "Enkelt innhold"},
	TG1008: {"Large print", "Storskrift"},
	TG1009: {"Readable print", "Leselig skrift"},
	TG1010: {"Braille text", "Blindeskrift"},
	TG1011: {"Sign language", "Tegnspråk"},
	TG1012: {"Tactile content", "Taktilt innhold"},
	TG1013: {"Bliss", "Bliss"},
	TG1014: {"Capital letters", "Store bokstaver"},
	TG1017: {"Widgit", "Widgit"},
}

var allAudiences = []Audience{
	TG1000, TG1001, TG1002, TG1003, TG1004, TG1005, TG1006, TG1007, TG1008,
	TG1009, TG1010, TG1011, TG1012, TG1013, TG1014, TG1015, TG1016, TG1017,
}

var audienceAliasesNo = map[Audience][]string{
	TG1010: {"Punktskrift"},
	TG1014: {"Versaler"},
}

var audienceAliasesEn = map[Audience][]string{
	TG1006: {"Easy-to-read text"},
}

var audienceCodes = map[string]Audience{
	"aa": TG1000,
	"a":  TG1001,
	"b":  TG1002,
	"bu": TG1003,
	"u":  TG1004,
	"mu": TG1005,
	"vu": TG1015,
	"v":  TG1016,
	"te": TG1010,
	"th": TG1013,
	"tb": TG1007,
	"td": TG1009,
	"ta": TG1006,
	"tj": TG1014,
	"tc": TG1008,
	"tg": TG1012,
	"tf": TG1011,
	"tk": TG1017,
}

func (a Audience) Code() string {
	return string(a)
}

func (a Audience) URL() string {
	return "https://schema.nb.no/Bibliographic/Values/" + string(a)
}

func (a Audience) Label(tag language.Tag) string {
	match, _, _ := localizer.Matcher.Match(tag)
	if match == language.Norwegian && audienceLabels[a][1] != "" {
		return audienceLabels[a][1]
	}
	return audienceLabels[a][0]
}

func (a Audience) Alias(tag language.Tag) []string {
	match, _, _ := localizer.Matcher.Match(tag)
	switch match {
	case language.Norwegian:
		return audienceAliasesNo[a]
	case language.English:
		return audienceAliasesEn[a]
	default:
		return nil
	}
}

func ParseAudience(s string) (Audience, error) {
	if a := Audience(s); a.Label(language.English) != "" {
		return a, nil
	}
	return "", ErrUnknown
}

func ParseAudienceURL(s string) (Audience, error) {
	return ParseAudience(strings.TrimPrefix(s, "https://schema.nb.no/Bibliographic/Values/"))
}

func ParseAudienceCode(s string) (Audience, error) {
	if a, ok := audienceCodes[s]; ok {
		return a, nil
	}
	return "", ErrUnknown
}

func AudienceOptions(lang language.Tag) (res [][2]string) {
	match, _, _ := localizer.Matcher.Match(lang)

	i := 0
	if match == language.Norwegian {
		i = 1

	}
	for _, a := range allAudiences {
		res = append(res, [2]string{string(a), audienceLabels[a][i]})
	}

	// Sort by label
	sort.Slice(res, func(i, j int) bool {
		return res[i][1] < res[j][1]
	})

	return res
}
