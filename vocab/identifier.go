package vocab

import "fmt"

var identifiers = map[string][2]string{
	"bibbi":         {"BIBBI", "https://id.bs.no/bibbi/%s"},
	"bibsys/aut":    {"BIBSYS", "https://authority.bibsys.no/authority/rest/authorities/v2/%s?format=xml"},
	"bibsys/pub":    {"BIBSYS", "https://marcpresentation.bibs.aws.unit.no/?mms_id=%s"},
	"gtin":          {"GTIN", ""},                         // Global Trade Item Number (formerly EAN: European Article Number)
	"isbn":          {"ISBN", ""},                         // International Standard Book Number
	"isni":          {"ISNI", "https://isni.org/isni/%s"}, // International Standard Name Identifier
	"issn":          {"ISSN", ""},                         // International Standard Serial Number
	"orcid":         {"ORCID", "http://orcid.org/%s"},     // Open Researcher and Contributor ID
	"snl":           {"Store norske leksikon", "https://snl.no/%s"},
	"viaf":          {"VIAF", "http://viaf.org/viaf/%s"}, // Virtual International Authority File
	"wikipedia/en":  {"Wikipedia (English)", "https://en.wikipedia.org/wiki/%s"},
	"wikipedia/no":  {"Wikipedia (norsk)", "https://no.wikipedia.org/wiki/%s"},
	"wikidata":      {"Wikidata", "https://www.wikidata.org/wiki/%s"},
	"worldcat":      {"WorldCat", "https://www.worldcat.org/identities/lccn-%s"},
	"nb/isbnforlag": {"Norske forlagsadresser", "https://nb.no/isbnforlag/record/%s"},
	"www":           {"WWW", "%s"},
	"isbn/prefix":   {"ISBN prefiks", ""},
	"nb/free":       {"Nasjonalbiblioteket (i det fri)", "https://urn.nb.no/%s"},
	"nb/norway":     {"Nasjonalbiblioteket (kun fra Norge)", "https://urn.nb.no/%s"},
	"nb/restricted": {"Nasjonalbiblioteket (begrenset)", "https://urn.nb.no/%s"},

	// TODO candidates:
	//discogs-artist  https://www.discogs.com/artist/32198
	//discogs-work    https://www.discogs.com/master/1282847
	//discogs-release https://www.discogs.com/release/895934
	//discogs-label   https://www.discogs.com/label/6785
}

// Identifier is an external, known Identifier.
type Identifier struct {
	Code  string
	Value string
	Label string
	URL   string // Optional
}

// ParseIdentifier creates an Identifier from the give code and value.
// If the code is not known, the Identifier Label will be set as the code,
// and the URL will be empty.
func ParseIdentifier(code, value string) Identifier {
	if id, found := identifiers[code]; found {
		var url string
		if id[1] != "" {
			url = fmt.Sprintf(id[1], value)
		}
		return Identifier{
			Code:  code,
			Value: value,
			Label: id[0],
			URL:   url,
		}
	}

	return Identifier{
		Code:  code,
		Value: value,
		Label: code,
	}
}
