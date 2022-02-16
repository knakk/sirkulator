package vocab

import "fmt"

var identifiers = map[string][2]string{
	"bibbi":    {"BIBBI", "https://id.bs.no/bibbi/%s"},
	"bibsys":   {"BIBSYS", "https://authority.bibsys.no/authority/rest/authorities/v2/%s?format=xml"},
	"gtin":     {"GTIN", ""},                         // Global Trade Item Number (formerly EAN: European Article Number)
	"isbn":     {"ISBN", ""},                         // International Standard Book Number
	"isni":     {"ISNI", "https://isni.org/isni/%s"}, // International Standard Name Identifier
	"issn":     {"ISSN", ""},                         // International Standard Serial Number
	"orcid":    {"ORCID", "http://orcid.org/%s"},     // Open Researcher and Contributor ID
	"snl":      {"Store norske leksikon", "https://snl.no/%s"},
	"viaf":     {"VIAF", "http://viaf.org/viaf/%s"}, // Virtual International Authority File
	"wiki-en":  {"Wikipedia (English)", "https://en.wikipedia.org/wiki"},
	"wiki-no":  {"Wikipedia (norsk)", "https://no.wikipedia.org/wiki"},
	"wikidata": {"Wikidata", "https://www.wikidata.org/wiki/%s"},
	"worldcat": {"WorldCat", "https://www.worldcat.org/identities/lccn-%s"},
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
