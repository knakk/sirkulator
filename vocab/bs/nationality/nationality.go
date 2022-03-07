package nationality

import (
	"github.com/knakk/sirkulator/internal/localizer"
	"golang.org/x/text/language"
)

// https://vokabular.bs.no/bibbi/nb/page/1189516

var bsNationalities = map[string][2]string{
	"aborig.":      {"Aboriginal", "Aboriginsk"}, // No equivalent MARC/ISO3166 country
	"afg.":         {"Afghan", "Afgansk"},
	"alb.":         {"Albanian", "Albansk"},
	"alg.":         {"Algerian", "Algerisk"},
	"am.":          {"American", "Amerikansk"}, // USA
	"angol.":       {"Angolian", "Angolsk"},
	"antig.":       {"Antiguan", "Antiguansk"},
	"arab.":        {"Arabic", "Arabisk"}, // No equivalent MARC/ISO3166 country
	"argen.":       {"Argentinian", "Argentinsk"},
	"arm.":         {"Armenian", "Armensk"},
	"aserb":        {"Azerbaijani", "Aserbajdsjansk"},
	"au.":          {"Australian", "Australsk"},
	"babyl.":       {"Babylonian", "Babylonsk"}, // No equivalent MARC/ISO3166 country
	"baham.":       {"Bahamian", "Bahamsk"},
	"bangl.":       {"Bangladeshian", "Bangladeshisk"},
	"barb.":        {"Barbadian", "Barbadisk"},
	"belg.":        {"Beligan", "Belgisk"},
	"benin.":       {"Beninian", "Beninsk"},
	"bhut.":        {"Bhutanian", "Bhutansk"},
	"guineab.":     {"Bissau-Guinean", "Bissauguineansk"},
	"boliv.":       {"Bolivian", "Boliviansk"},
	"bosn.":        {"Bosnian", "Bosnisk"},
	"botsw.":       {"Botswanian", "Botswansk"},
	"bras.":        {"Brazilian", "Brasilsk"},
	"kongol.braz.": {"Congo Brazzavillian", "Brazzavillekongolesisk"},
	"brun.":        {"Bruneian", "Bruneinsk"}, // ?
	"bulg.":        {"Bulgarian", "Bulgarsk"},
	"burkin.":      {"Burkinian", "Burkinsk"},
	"burund.":      {"Burundian", "Burundisk"},

	"n.":   {"Norwegian", "Norsk"},
	"sam.": {"Sami", "Samisk"}, // No equivalent MARC/ISO3166 country
}

var bsToMarc = map[string]string{
	"afg.":   "af",
	"alb.":   "aa",
	"alg.":   "ae",
	"am.":    "xxu",
	"angol.": "ao",
	"antig.": "aq",
	"argen.": "ag",
	"arm.":   "ai",
	"aserb":  "aj",
	"au.":    "at",
	"brun.":  "bx",
}

var bsToISO3166 = map[string]string{
	"n.": "NO",
}

func Options(lang language.Tag) (res [][2]string) {
	match, _, _ := localizer.Matcher.Match(lang)
	i := 0
	if match == language.Norwegian {
		i = 1
	}
	for k, v := range bsNationalities {
		res = append(res, [2]string{k, v[i]})
	}
	return res
}
