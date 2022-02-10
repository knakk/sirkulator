// Code generated by running "go generate" in golang.org/x/text. DO NOT EDIT.

package translations

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

type dictionary struct {
	index []uint32
	data  string
}

func (d *dictionary) Lookup(key string) (data string, ok bool) {
	p, ok := messageKeyToIndex[key]
	if !ok {
		return "", false
	}
	start, end := d.index[p], d.index[p+1]
	if start == end {
		return "", false
	}
	return d.data[start:end], true
}

func init() {
	dict := map[string]catalog.Dictionary{
		"en": &dictionary{index: enIndex, data: enData},
		"no": &dictionary{index: noIndex, data: noData},
	}
	fallback := language.MustParse("en")
	cat, err := catalog.NewFromMap(dict, catalog.Fallback(fallback))
	if err != nil {
		panic(err)
	}
	message.DefaultCatalog = cat
}

var messageKeyToIndex = map[string]int{
	"%d imported OK!":          15,
	"1 per line":               12,
	"Circulation":              1,
	"Configuration":            5,
	"Holdings":                 4,
	"Home":                     0,
	"ISBN, ISSN or EAN":        11,
	"Identifiers":              10,
	"Import":                   9,
	"Metadata":                 3,
	"Orders":                   2,
	"Preview":                  13,
	"Search/browse catalogue":  8,
	"Show metadata for review": 7,
	"Show recent transactions": 6,
	"wait...":                  14,
}

var enIndex = []uint32{ // 17 elements
	0x00000000, 0x00000005, 0x00000011, 0x00000018,
	0x00000021, 0x0000002a, 0x00000038, 0x00000051,
	0x0000006a, 0x00000082, 0x00000089, 0x00000095,
	0x000000a7, 0x000000b2, 0x000000ba, 0x000000c2,
	0x000000d5,
} // Size: 92 bytes

const enData string = "" + // Size: 213 bytes
	"\x02Home\x02Circulation\x02Orders\x02Metadata\x02Holdings\x02Configurati" +
	"on\x02Show recent transactions\x02Show metadata for review\x02Search/bro" +
	"wse catalogue\x02Import\x02Identifiers\x02ISBN, ISSN or EAN\x021 per lin" +
	"e\x02Preview\x02wait...\x02%[1]d imported OK!"

var noIndex = []uint32{ // 17 elements
	0x00000000, 0x00000005, 0x00000011, 0x0000001e,
	0x00000027, 0x0000002f, 0x0000003d, 0x00000055,
	0x00000075, 0x0000008a, 0x00000093, 0x000000a3,
	0x000000b8, 0x000000c4, 0x000000d1, 0x000000d9,
	0x000000ed,
} // Size: 92 bytes

const noData string = "" + // Size: 237 bytes
	"\x02Hjem\x02Sirkulasjon\x02Bestillinger\x02Metadata\x02Bestand\x02Konfig" +
	"urasjon\x02Vis siste transaksjoner\x02Vis opplysninger til gjennomsyn" +
	"\x02Søk/bla i katalogen\x02Importer\x02Identifikatorer\x02ISBN, ISSN ell" +
	"er EAN\x021 per linje\x02Forhåndsvis\x02vent...\x02%[1]d importert OK!"

	// Total table size 634 bytes (0KiB); checksum: C0548E06
