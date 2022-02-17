package marc

import (
	"errors"

	"github.com/knakk/sirkulator/internal/localizer"
	"golang.org/x/text/language"
)

// The list of relators is manually generated from:
// 	https://katalogisering.bibsys.no/files/2019/04/Liste_over_relasjoner.pdf
// which was last updated 2019-03-06
// The original source of English codes is maintained by Library of Congress:
// 	https://www.loc.gov/marc/relators/relaterm.html
// TODO also check https://rdakatalogisering.unit.no/files/2021/12/Liste_over_relasjoner_RDA.pdf
var relators = map[string][2]string{
	"act": {"Actor", "skuespiller"},
	"adp": {"Adapter", "bearbeider"},
	"aft": {"Author of afterword, colophon, etc.", "forfatter (etterord)"},
	"anm": {"Animator", "tegnefilmkunstner"},
	"arc": {"Architect", "arkitekt"},
	"arr": {"Arranger", "arrangør (musikk)"},
	"art": {"Artist", "bildende kunstner"},
	"aui": {"Author of introduction", "forfatter (forord)"},
	"aus": {"Author of screenplay", "manusforfatter (filmmanus)"},
	"aut": {"Author", "forfatter"}, // Brukes for ansvarlig for tekstlig innhold. Når hovedansvarlig for et verk ikke er ansvarlig for tekst, brukes den kode som beskriver relasjonen.
	"bjd": {"Bookjacket designer", "omslagsdesigner"},
	"ccp": {"Conceptor", "idéskaper"},
	"chr": {"Choreographer", "koreograf"},
	"cmm": {"Commentator", "kommentator"},
	"cmp": {"Composer", "komponist"},
	"cng": {"Cinematographer", "filmfotograf"},
	"cnd": {"Conductor", "dirigent"},
	"cph": {"Copyright holder", "rettighetshaver"},
	"cre": {"Creator", "skaper"},
	"crp": {"Correspondent", "brevskriver"},
	"ctb": {"Contributor", "bidragsyter"},
	"ctg": {"Cartographer", "kartograf"},
	"cur": {"Curator", "kurator (utstillinger)"},
	"dnc": {"Dancer", "danser"},
	"drt": {"Director", "regissør/instruktør"},
	"dst": {"Distributor", "distributør"},
	"dte": {"Dedicatee", "person tilegnet"},
	"dgg": {"Degree grantor", "eksamenssted"},
	"dub": {"Dubious author", "usikkert forfatterskap"},
	"edt": {"Editor", "redaktør"},
	"flm": {"Film editor", "klipper"},
	"his": {"Host institution", "vertsinstitusjon"},
	"hst": {"Host", "programleder"},
	"ill": {"Illustrator", "illustratør"},
	"itr": {"Instrumentalist", "instrumentalist"},
	"ive": {"Interviewee", "intervjuobjekt"},
	"ivr": {"Interviewer", "intervjuer"},
	"lbt": {"Librettist", "librettoforfatter"},
	"lgd": {"Lighting designer", "lysmester"},
	"lyr": {"Lyricist", "tekstforfatter (musikk)"}, // Brukes for forfatter av sangtekst. For forfatter av lyrisk tekst som senere er blitt tonesatt brukes kode for forfatter [aut].
	"mus": {"Musician", "musiker"},
	"nrt": {"Narrator", "forteller"}, // Brukes bl.a. for innleser av lydbok.
	"orm": {"Organizer", "arrangør"},
	"oth": {"Other", "annet"},
	"own": {"Owner", "eier"},
	"pat": {"Patron", "sponsor"},
	"pbd": {"Publishing director", "forlagsredaktør"},
	"pbl": {"Publisher", "forlag/utgiver"},
	"pht": {"Photographer", "fotograf"},
	"prf": {"Performer", "utøver"},
	"prg": {"Programmer", "programmerer"},
	"pro": {"Producer", "produsent"},
	"prt": {"Printer", "trykker"},
	"rcp": {"Recipient", "mottaker (korrespondanse)"},
	"res": {"Researcher", "forsker"},
	"rev": {"Reviewer", "anmelder"},
	"scl": {"Sculptor", "skulptør"},
	"sng": {"Singer", "sanger"},
	"std": {"Set designer", "scenograf"},
	"stl": {"Storyteller", "forteller"}, // Bruk Narrator [nrt]
	"trl": {"Translator", "oversetter"},
	"voc": {"Vocalist", "vokalist"}, // Bruk Singer [sng]
	"msd": {"Musical director", "musikalsk leder"},
	"pmn": {"Production manager", "produksjonsleder"},
	"pdr": {"Project director", "prosjektleder"},
	"tcd": {"Technical director", "teknisk leder"},
}

// Relator is a known Marc relator, associated with a 3-letter code.
// The codes are maintained by Library of Congress.
type Relator struct {
	code string
}

// ParseRelator parses the given string and returns a Relator if
// it matches a known 3-letter Marc Relator code.
func ParseRelator(s string) (Relator, error) {
	if _, ok := relators[s]; ok {
		return Relator{code: s}, nil
	}
	return Relator{}, errors.New("marc: unknown relator")
}

// Code returns the Marc code for the Realtor.
func (r Relator) Code() string {
	return r.code
}

// Label returns a string representation of the Marc relator in the desired language.
// Only Norwegian and English are currently supported.
func (r Relator) Label(tag language.Tag) string {
	match, _, _ := localizer.Matcher.Match(tag)
	if match == language.Norwegian {
		return relators[r.code][1]
	}
	return relators[r.code][0]
}
