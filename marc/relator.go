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
	"act": {"Actor", "Skuespiller"},
	"adp": {"Adapter", "Bearbeider"},
	"aft": {"Author of afterword, colophon, etc.", "Forfatter (etterord)"},
	"anm": {"Animator", "Tegnefilmkunstner"},
	"arc": {"Architect", "Arkitekt"},
	"arr": {"Arranger", "Arrangør (musikk)"},
	"art": {"Artist", "Bildende kunstner"},
	"aui": {"Author of introduction", "Forfatter (forord)"},
	"aus": {"Author of screenplay", "Manusforfatter (filmmanus)"},
	"aut": {"Author", "Forfatter"}, // Brukes for ansvarlig for tekstlig innhold. Når hovedansvarlig for et verk ikke er ansvarlig for tekst, brukes den kode som beskriver relasjonen.
	"bjd": {"Bookjacket designer", "Omslagsdesigner"},
	"ccp": {"Conceptor", "Idéskaper"},
	"chr": {"Choreographer", "Koreograf"},
	"cmm": {"Commentator", "Kommentator"},
	"cmp": {"Composer", "Komponist"},
	"cng": {"Cinematographer", "Filmfotograf"},
	"cnd": {"Conductor", "Dirigent"},
	"cph": {"Copyright holder", "Rettighetshaver"},
	"cre": {"Creator", "Skaper"},
	"crp": {"Correspondent", "Brevskriver"},
	"ctb": {"Contributor", "Bidragsyter"},
	"ctg": {"Cartographer", "Kartograf"},
	"cur": {"Curator", "Kurator (utstillinger)"},
	"dnc": {"Dancer", "Danser"},
	"drt": {"Director", "Regissør/instruktør"},
	"dst": {"Distributor", "Distributør"},
	"dte": {"Dedicatee", "Person tilegnet"},
	"dgg": {"Degree grantor", "Eksamenssted"},
	"dub": {"Dubious author", "Usikkert forfatterskap"},
	"edt": {"Editor", "Redaktør"},
	"flm": {"Film editor", "Klipper"},
	"his": {"Host institution", "Vertsinstitusjon"},
	"hst": {"Host", "Programleder"},
	"ill": {"Illustrator", "Illustratør"},
	"itr": {"Instrumentalist", "Instrumentalist"},
	"ive": {"Interviewee", "Intervjuobjekt"},
	"ivr": {"Interviewer", "Intervjuer"},
	"lbt": {"Librettist", "Librettoforfatter"},
	"lgd": {"Lighting designer", "Lysmester"},
	"lyr": {"Lyricist", "Tekstforfatter (musikk)"}, // Brukes for forfatter av sangtekst. For forfatter av lyrisk tekst som senere er blitt tonesatt brukes kode for forfatter [aut].
	"mus": {"Musician", "Musiker"},
	"nrt": {"Narrator", "Forteller"}, // Brukes bl.a. for innleser av lydbok.
	"orm": {"Organizer", "Arrangør"},
	"oth": {"Other", "Annet"},
	"own": {"Owner", "Eier"},
	"pat": {"Patron", "Sponsor"},
	"pbd": {"Publishing director", "Forlagsredaktør"},
	"pbl": {"Publisher", "Forlag/utgiver"},
	"pht": {"Photographer", "Fotograf"},
	"prf": {"Performer", "Utøver"},
	"prg": {"Programmer", "Programmerer"},
	"pro": {"Producer", "Produsent"},
	"prt": {"Printer", "Trykker"},
	"rcp": {"Recipient", "Mottaker (korrespondanse)"},
	"res": {"Researcher", "Forsker"},
	"rev": {"Reviewer", "Anmelder"},
	"scl": {"Sculptor", "Skulptør"},
	"sng": {"Singer", "Sanger"},
	"std": {"Set designer", "Scenograf"},
	"stl": {"Storyteller", "Forteller"}, // Bruk Narrator [nrt]
	"trl": {"Translator", "Oversetter"},
	"voc": {"Vocalist", "Vokalist"}, // Bruk Singer [sng]
	"msd": {"Musical director", "Musikalsk leder"},
	"pmn": {"Production manager", "Produksjonsleder"},
	"pdr": {"Project director", "Prosjektleder"},
	"tcd": {"Technical director", "Teknisk leder"},
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
