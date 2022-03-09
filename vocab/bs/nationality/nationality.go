package nationality

import (
	"sort"
	"strings"

	"github.com/knakk/sirkulator/internal/localizer"
	"github.com/knakk/sirkulator/vocab"
	"golang.org/x/text/language"
)

// Nationality is a nationality/regional group (Nasjonalitet/regional gruppe) from
// the Bibbi authority register maintained by Biblioeksentralen SA.
// https://vokabular.bs.no/bibbi/nb/
type Nationality string

const (
	Aborig     Nationality = "aborig"
	Afg        Nationality = "afg"
	Alb        Nationality = "alb"
	Alg        Nationality = "alg"
	Am         Nationality = "am"
	Angol      Nationality = "angol"
	Antig      Nationality = "antig"
	Arab       Nationality = "arab"
	Argen      Nationality = "argen"
	Arm        Nationality = "arm"
	Aserb      Nationality = "aserb"
	Au         Nationality = "au"
	Babyl      Nationality = "babyl"
	Baham      Nationality = "baham"
	Bangl      Nationality = "bangl"
	Barb       Nationality = "barb"
	Belg       Nationality = "belg"
	Benin      Nationality = "benin"
	Bhut       Nationality = "bhut"
	Guineab    Nationality = "guineab"
	Boliv      Nationality = "boliv"
	Bosn       Nationality = "bosn"
	Botsw      Nationality = "botsw"
	Bras       Nationality = "bras"
	KongolBraz Nationality = "kongol.braz"
	Brun       Nationality = "brun"
	Bulg       Nationality = "bulg"
	Burkin     Nationality = "burkin"
	Burund     Nationality = "burund"
	Chil       Nationality = "chil"
	Colomb     Nationality = "colomb"
	Costaric   Nationality = "costaric"
	Cub        Nationality = "cub"
	D          Nationality = "d"
	Dominik    Nationality = "dominik"
	Ecuad      Nationality = "ecuad"
	Egypt      Nationality = "egypt"
	Emiratarab Nationality = "emiratarab"
	Eng        Nationality = "eng"
	Eritr      Nationality = "eritr"
	Est        Nationality = "est"
	Esw        Nationality = "esw"
	Etiop      Nationality = "etiop"
	Filip      Nationality = "filip"
	Fi         Nationality = "fi"
	Fr         Nationality = "fr"
	Fær        Nationality = "fær"
	Gabon      Nationality = "gabon"
	Gamb       Nationality = "gamb"
	Gas        Nationality = "gas"
	Georg      Nationality = "georg"
	Ghan       Nationality = "ghan"
	Gr         Nationality = "gr"
	Grenad     Nationality = "grenad"
	Grønl      Nationality = "grønl"
	Guadel     Nationality = "guadel"
	Guatem     Nationality = "guatem"
	Guin       Nationality = "guin"
	Guyan      Nationality = "guyan"
	Hait       Nationality = "hait"
	Hond       Nationality = "hond"
	Hviter     Nationality = "hviter"
	Ind        Nationality = "ind"
	Indon      Nationality = "indon"
	Irak       Nationality = "irak"
	Iran       Nationality = "iran"
	Ir         Nationality = "ir"
	Isl        Nationality = "isl"
	Isr        Nationality = "isr"
	It         Nationality = "it"
	Ivor       Nationality = "ivor"
	Jam        Nationality = "jam"
	Jap        Nationality = "jap"
	Jemen      Nationality = "jemen"
	Jord       Nationality = "jord"
	Jug        Nationality = "jug"
	Kamb       Nationality = "kamb"
	Kamer      Nationality = "kamer"
	Kan        Nationality = "kan"
	Kappverd   Nationality = "kappverd"
	Kas        Nationality = "kas"
	Katal      Nationality = "katal"
	Ken        Nationality = "ken"
	Kin        Nationality = "kin"
	Kirg       Nationality = "kirg"
	Komor      Nationality = "komor"
	Kongol     Nationality = "kongol"
	Kor        Nationality = "kor"
	Kos        Nationality = "kos"
	Kroat      Nationality = "kroat"
	Kurd       Nationality = "kurd"
	Kuw        Nationality = "kuw"
	Kypr       Nationality = "kypr"
	Laot       Nationality = "laot"
	Latv       Nationality = "latv"
	Lesot      Nationality = "lesot"
	Liban      Nationality = "liban"
	Liber      Nationality = "liber"
	Liby       Nationality = "liby"
	Liecht     Nationality = "liecht"
	Lit        Nationality = "lit"
	Lux        Nationality = "lux"
	Mak        Nationality = "mak"
	Malaw      Nationality = "malaw"
	Malay      Nationality = "malay"
	Mali       Nationality = "mali"
	Malt       Nationality = "malt"
	Maori      Nationality = "maori"
	Marok      Nationality = "marok"
	Maurit     Nationality = "maurit"
	Mauret     Nationality = "mauret"
	Mex        Nationality = "mex"
	Mold       Nationality = "mold"
	Mong       Nationality = "mong"
	Montenegr  Nationality = "montenegr"
	Mosamb     Nationality = "mosamb"
	Myanm      Nationality = "myanm"
	Namib      Nationality = "namib"
	Ned        Nationality = "ned"
	Nep        Nationality = "nep"
	Newzeal    Nationality = "newzeal"
	Nicarag    Nationality = "nicarag"
	Nig        Nationality = "nig"
	Niger      Nationality = "niger"
	Nordir     Nationality = "nordir"
	Nordkor    Nationality = "nordkor"
	N          Nationality = "n"
	Pak        Nationality = "pak"
	Pal        Nationality = "pal"
	Panam      Nationality = "panam"
	Pap        Nationality = "pap"
	Parag      Nationality = "parag"
	Pers       Nationality = "pers"
	Peru       Nationality = "peru"
	Pol        Nationality = "pol"
	Portug     Nationality = "portug"
	Puert      Nationality = "puert"
	Qat        Nationality = "qat"
	Rom        Nationality = "rom"
	Rum        Nationality = "rum"
	R          Nationality = "r"
	Rwand      Nationality = "rwand"
	Salvad     Nationality = "salvad"
	Sam        Nationality = "sam"
	Samoan     Nationality = "samoan"
	Sanktluc   Nationality = "sanktluc"
	Saudiarab  Nationality = "saudiarab"
	Senegal    Nationality = "senegal"
	Serb       Nationality = "serb"
	Sey        Nationality = "sey"
	Sierral    Nationality = "sierral"
	Singapor   Nationality = "singapor"
	Sk         Nationality = "sk"
	Slovak     Nationality = "slovak"
	Sloven     Nationality = "sloven"
	Somal      Nationality = "somal"
	Sp         Nationality = "sp"
	Srilank    Nationality = "srilank"
	Sudan      Nationality = "sudan"
	Surin      Nationality = "surin"
	Sveits     Nationality = "sveits"
	Sv         Nationality = "sv"
	Syr        Nationality = "syr"
	Sørafr     Nationality = "sørafr"
	Sørkor     Nationality = "sørkor"
	Sørsudan   Nationality = "sørsudan"
	Tadsj      Nationality = "tadsj"
	Tahit      Nationality = "tahit"
	Taiw       Nationality = "taiw"
	Tanz       Nationality = "tanz"
	Thai       Nationality = "thai"
	Tib        Nationality = "tib"
	Togo       Nationality = "togo"
	Trinid     Nationality = "trinid"
	Tchad      Nationality = "tchad"
	Tsj        Nationality = "tsj"
	Tsjet      Nationality = "tsjet"
	Tun        Nationality = "tun"
	Turkm      Nationality = "turkm"
	Tyrk       Nationality = "tyrk"
	T          Nationality = "t"
	Ugand      Nationality = "ugand"
	Ukr        Nationality = "ukr"
	Ung        Nationality = "ung"
	Urug       Nationality = "urug"
	Usb        Nationality = "usb"
	Venez      Nationality = "venez"
	Viet       Nationality = "viet"
	Wal        Nationality = "wal"
	Zair       Nationality = "zair"
	Zamb       Nationality = "zamb"
	Zimb       Nationality = "zimb"
	Øst        Nationality = "øst"
)

var allNationalities = []Nationality{
	Aborig, Afg, Alb, Alg, Am, Angol, Antig, Arab, Argen, Arm, Aserb, Au, Babyl, Baham, Bangl,
	Barb, Belg, Benin, Bhut, Guineab, Boliv, Bosn, Botsw, Bras, KongolBraz, Brun, Bulg, Burkin,
	Burund, Chil, Colomb, Costaric, Cub, D, Dominik, Ecuad, Egypt, Emiratarab, Eng, Eritr, Est,
	Esw, Etiop, Filip, Fi, Fr, Fær, Gabon, Gamb, Gas, Georg, Ghan, Gr, Grenad, Grønl, Guadel,
	Guatem, Guin, Guyan, Hait, Hond, Hviter, Ind, Indon, Irak, Iran, Ir, Isl, Isr, It, Ivor,
	Jam, Jap, Jemen, Jord, Jug, Kamb, Kamer, Kan, Kappverd, Kas, Katal, Ken, Kin, Kirg, Komor,
	Kongol, Kor, Kos, Kroat, Kurd, Kuw, Kypr, Laot, Latv, Lesot, Liban, Liber, Liby, Liecht,
	Lit, Lux, Mak, Malaw, Malay, Mali, Malt, Maori, Marok, Maurit, Mauret, Mex, Mold, Mong,
	Montenegr, Mosamb, Myanm, Namib, Ned, Nep, Newzeal, Nicarag, Nig, Niger, Nordir, Nordkor,
	N, Pak, Pal, Panam, Pap, Parag, Pers, Peru, Pol, Portug, Puert, Qat, Rom, Rum, R, Rwand,
	Salvad, Sam, Samoan, Sanktluc, Saudiarab, Senegal, Serb, Sey, Sierral, Singapor, Sk,
	Slovak, Sloven, Somal, Sp, Srilank, Sudan, Surin, Sveits, Sv, Syr, Sørafr, Sørkor, Sørsudan,
	Tadsj, Tahit, Taiw, Tanz, Thai, Tib, Togo, Trinid, Tchad, Tsj, Tsjet, Tun, Turkm, Tyrk, T,
	Ugand, Ukr, Ung, Urug, Usb, Venez, Viet, Wal, Zair, Zamb, Zimb, Øst,
}

func (n Nationality) URI() string {
	return "bs/" + string(n)
}

func Parse(s string) (Nationality, error) {
	s = strings.TrimSuffix(s, ".")
	if _, ok := labels[Nationality(s)]; ok {
		return Nationality(s), nil
	}
	return "", vocab.ErrUnknown
}

// TODO verify english labels against wikipedia:
// https://en.wikipedia.org/wiki/List_of_adjectival_and_demonymic_forms_for_countries_and_nations
var labels = map[Nationality][2]string{
	Aborig:     {"Aboriginal", "Aboriginsk"}, // No equivalent MARC/ISO3166 country
	Afg:        {"Afghan", "Afgansk"},
	Alb:        {"Albanian", "Albansk"},
	Alg:        {"Algerian", "Algerisk"},
	Am:         {"American", "Amerikansk"}, // USA
	Angol:      {"Angolian", "Angolsk"},
	Antig:      {"Antiguan", "Antiguansk"},
	Arab:       {"Arabic", "Arabisk"}, // No equivalent MARC/ISO3166 country
	Argen:      {"Argentinian", "Argentinsk"},
	Arm:        {"Armenian", "Armensk"},
	Aserb:      {"Azerbaijani", "Aserbajdsjansk"},
	Au:         {"Australian", "Australsk"},
	Babyl:      {"Babylonian", "Babylonsk"}, // No equivalent MARC/ISO3166 country
	Baham:      {"Bahamian", "Bahamsk"},
	Bangl:      {"Bangladeshian", "Bangladeshisk"},
	Barb:       {"Barbadian", "Barbadisk"},
	Belg:       {"Beligan", "Belgisk"},
	Benin:      {"Beninian", "Beninsk"},
	Bhut:       {"Bhutanian", "Bhutansk"},
	Guineab:    {"Bissau-Guinean", "Bissauguineansk"},
	Boliv:      {"Bolivian", "Boliviansk"},
	Bosn:       {"Bosnian", "Bosnisk"},
	Botsw:      {"Botswanian", "Botswansk"},
	Bras:       {"Brazilian", "Brasilsk"},
	KongolBraz: {"Congo Brazzavillian", "Brazzavillekongolesisk"},
	Brun:       {"Bruneian", "Bruneinsk"}, // ?
	Bulg:       {"Bulgarian", "Bulgarsk"},
	Burkin:     {"Burkinian", "Burkinsk"},
	Burund:     {"Burundian", "Burundisk"},
	Chil:       {"Chilean", "Chilensk"},
	Colomb:     {"Columbian", "Columbiansk"},
	Costaric:   {"Costa Rican", "Costarikansk"},
	Cub:        {"Cuban", "Kubansk"},
	D:          {"Danish", "Dansk"},
	Dominik:    {"Dominican", "Dominikansk"},
	Ecuad:      {"Ecuadorian", "Ecuadoriansk"},
	Egypt:      {"Egyptian", "Egyptisk"},
	Emiratarab: {"Emiratarabian", "Emiratarabisk"}, // ?
	Eng:        {"English", "Engelsk"},
	Eritr:      {"Eritrean", "Eritreisk"},
	Est:        {"Estonian", "Estlandsk"},
	Esw:        {"Eswantian", "Eswantisk"}, // ? (former swaziland)
	Etiop:      {"Ethiopian", "Etiopisk"},
	Filip:      {"Philipinian", "Filipinsk"},
	Fi:         {"Finnish", "Finsk"},
	Fr:         {"French", "Fransk"},
	Fær:        {"Faroese", "Færøysk"},
	Gabon:      {"Gabonian", "Gabonsk"}, // ?
	Gamb:       {"Gambian", "Gambisk"},
	Gas:        {"Malagasy", "Gassisk"},
	Georg:      {"Georgian", "Georgisk"},
	Ghan:       {"Ghanesian", "Ghanesisk"},
	Gr:         {"Greek", "Gresk"},
	Grenad:     {"Grenadian", "Grenadisk"}, // ?
	Grønl:      {"Greenlandic", "Grønlandsk"},
	Guadel:     {"Guadeloupe", "Guadeloupeisk"}, // ?
	Guatem:     {"Guatemalan", "Guatemalansk"},  // ?
	Guin:       {"Guinean", "Guineansk"},
	Guyan:      {"Guyanese", "Guyanansk"}, // ?
	Hait:       {"Haitian", "Haitisk"},
	Hond:       {"Honduran", "Honduransk"},
	Hviter:     {"Belarusian", "Hviterussisk"},
	Ind:        {"Indian", "Indisk"},
	Indon:      {"Indonesian", "Indonesisk"},
	Irak:       {"Iraqi", "Irakisk"},
	Iran:       {"Iranian", "Iransk"},
	Ir:         {"Irish", "Irsk"},
	Isl:        {"Icelandic", "Islandsk"},
	Isr:        {"Israeli", "Israelsk"},
	It:         {"Italian", "Italiensk"},
	Ivor:       {"Ivorian", "Ivoriansk"},
	Jam:        {"Jamaican", "Jamaikansk"},
	Jap:        {"Japanese", "Japansk"},
	Jemen:      {"Yemeni", "Jemenittisk"},
	Jord:       {"Jordanian", "Jordansk"},
	Jug:        {"Yugoslavian", "Jugoslavisk"},
	Kamb:       {"Cambodian", "Kambodsjansk"},
	Kamer:      {"Cameroonian", "Kamerunsk"},
	Kan:        {"Canadian", "Kanadisk"},
	Kappverd:   {"Cabo Verdean", "Kappverdisk"}, //?
	Kas:        {"Kazakh", "Kazakhstansk"},
	Katal:      {"Katalan", "Katalansk"}, // No equivalent MARC/ISO3166 country
	Ken:        {"Kenyan", "Kenyansk"},
	Kin:        {"Chinese", "Kinesisk"},
	Kirg:       {"Kirgiz", "Kirgisisk"},
	Komor:      {"Comoran", "Komorisk"}, // ?
	Kongol:     {"Congolese", "Kongolesisk"},
	Kor:        {"Corean", "Koreansk"},
	Kos:        {"Kosovar", "Kosovarsk"}, // ??
	Kroat:      {"Croatian", "Kroatisk"},
	Kurd:       {"Kurdish", "Kurdisk"},
	Kuw:        {"Kuwaiti", "Kuwaitisk"},
	Kypr:       {"Cypriot", "Kypriotisk"},
	Laot:       {"Lao", "Laotisk"},
	Latv:       {"Latvian", "Latvisk"},
	Lesot:      {"Basotho", "Lesothisk"},
	Liban:      {"Lebanese", "Libanesisk"},
	Liber:      {"Liberian", "Liberisk"},
	Liby:       {"Libyan", "Libysk"},
	Liecht:     {"Liechtensteiner", "Liechtensteinsk"},
	Lit:        {"Lithuanian", "Litausk"},
	Lux:        {"Luxembourg", "Luxemburgsk"},
	Mak:        {"Macedonian", "Makedonsk"},
	Malaw:      {"Malawian", "Malawisk"},
	Malay:      {"Malaysian", "Malaysisk"},
	Mali:       {"Malian", "Malisk"},
	Malt:       {"Maltese", "Maltesisk"},
	Maori:      {"Maori", "Maorisk"},
	Marok:      {"Moroccan", "Maokansk"},
	Maurit:     {"Mauritian", "Mauritisk"},
	Mauret:     {"Mauritanian", "Mauritansk"},
	Mex:        {"Mexican", "Meksikansk"},
	Mold:       {"Moldovian", "Moldovisk"},
	Mong:       {"Monglian", "Mongolsk"},
	Montenegr:  {"Montenegrin", "Montenegrinsk"},
	Mosamb:     {"Mozambican", "Mosambisk"},
	Myanm:      {"Myanma", "Myanmarsk"},
	Namib:      {"Namibian", "Namibisk"},
	Ned:        {"Dutch", "Nederlandsk"},
	Nep:        {"Nepalese", "Nepalsk"},
	Newzeal:    {"New Zeland", "Ny-zelandsk"},
	Nicarag:    {"Nicaraguan", "Nicaraguansk"},
	Nig:        {"Nigerien", "Nigerisk"}, // ?
	Niger:      {"Nigerian", "Nigeriansk"},
	Nordir:     {"Northern Irish", "Nord-irsk"},
	Nordkor:    {"North Korean", "Nord-koreansk"},
	N:          {"Norwegian", "Norsk"},
	Pak:        {"Pakistani", "Pakistansk"},
	Pal:        {"Palestinian", "Palestninsk"},
	Panam:      {"Panamanian", "Panamansk"}, // ?
	Pap:        {"Papuan", "Papuansk"},
	Parag:      {"Paraguayan", "Paraguayansk"},
	Pers:       {"Persian", "Persisk"},
	Peru:       {"Peruvian", "Peruviansk"},
	Pol:        {"Polish", "Polsk"},
	Portug:     {"Portuguese", "Portugisisk"},
	Puert:      {"Puerto Rican", "Puertorikisk"},
	Qat:        {"Qatari", "Qatarsk"},
	Rom:        {"Roman", "Romersk"}, // No equivalent MARC/ISO3166 country
	Rum:        {"Romanian", "Rumensk"},
	R:          {"Russian", "Russisk"},
	Rwand:      {"Rwandan", "Rwandisk"},
	Salvad:     {"Salvadoran", "Salvadorisk"}, // ?
	Sam:        {"Sami", "Samisk"},            // No equivalent MARC/ISO3166 country
	Samoan:     {"Samoan", "Samoisk"},
	Sanktluc:   {"Saint Lucian", "Sanktalusisk"}, // ?
	Saudiarab:  {"Saudi", "Saudi-arabisk"},
	Senegal:    {"Senegalese", "Senegalesisk"},
	Serb:       {"Serbian", "Serbisk"},
	Sey:        {"Seychellois", "Seychellisk"},
	Sierral:    {"Sierra Leonean", "Sierraleonsk"},
	Singapor:   {"Singapore", "Singaporsk"},
	Sk:         {"Scottish", "Skotsk"},
	Slovak:     {"Slovakian", "Slovakisk"},
	Sloven:     {"Slovenian", "Slovensk"},
	Somal:      {"Somali", "Somalsk"},
	Sp:         {"Spanish", "Spansk"},
	Srilank:    {"Sri Lankan", "Srilankesisk"},
	Sudan:      {"Sudanese", "Sudansk"},
	Surin:      {"Surinamese", "Surinamsk"},
	Sveits:     {"Swiss", "Sveitsisk"},
	Sv:         {"Swedish", "Svensk"},
	Syr:        {"Syrian", "Syrisk"},
	Sørafr:     {"South African", "Sør-afrikansk"},
	Sørkor:     {"South Korean", "Sør-koreansk"},
	Sørsudan:   {"Sout Sudanese", "Sør-sudansk"},
	Tadsj:      {"Tajikistani", "Tadjikisk"},
	Tahit:      {"Tahitian", "Tahitisk"},
	Taiw:       {"Taiwanese", "Taiwansk"},
	Tanz:       {"Tanzanian", "Tanzaniansk"},
	Thai:       {"Thai", "Thailandsk"},
	Tib:        {"Tibetan", "Tibetansk"},
	Togo:       {"Togolese", "Togolesisk"}, // ?
	Trinid:     {"Trinidadian", "Trinidadsk"},
	Tchad:      {"Chadian", "Tsjadisk"},
	Tsj:        {"Czech", "Tsjekkisk"},
	Tsjet:      {"Chechen", "Tsjetsjensk"},
	Tun:        {"Tunisian", "Tunisisk"},
	Turkm:      {"Turkmen", "Turkmensk"},
	Tyrk:       {"Turkish", "Tyrkisk"},
	T:          {"German", "Tysk"},
	Ugand:      {"Ugandan", "Ugandisk"},
	Ukr:        {"Ukranian", "Ukrainsk"},
	Ung:        {"Hungarian", "Ungarsk"},
	Urug:       {"Uruguayan", "Uruguaysk"},
	Usb:        {"Uzbek", "Usbekisk"},
	Venez:      {"Venezuelan", "Venezuelansk"},
	Viet:       {"Vitenamese", "Vietnamsk"},
	Wal:        {"Walisian", "Walisisk"},
	Zair:       {"Zairian", "Zairisk"}, // ?
	Zamb:       {"Zambian", "Zambisk"},
	Zimb:       {"Zimbabwean", "Zimbabwisk"}, // ?
	Øst:        {"Austrian", "Østeriksk"},
}

/*
var bsToMarc = map[Nationality]string{
	Afg:   "af",
	Alb:   "aa",
	Alg:   "ae",
	Am:    "xxu",
	Angol: "ao",
	Antig: "aq",
	Argen: "ag",
	Arm:   "ai",
	Aserb: "aj",
	Au:    "at",
	Brun:  "bx",
	// etc
}

var bsToISO3166 = map[Nationality]string{
	N: "NO",
	// etc
}
*/

func Options(lang language.Tag) (res [][2]string) {
	match, _, _ := localizer.Matcher.Match(lang)
	i := 0
	if match == language.Norwegian {
		i = 1
	}
	for _, n := range allNationalities {
		res = append(res, [2]string{string(n), labels[n][i]})
	}

	// Sort by label
	sort.Slice(res, func(i, j int) bool {
		return res[i][1] < res[j][1]
	})

	return res
}
