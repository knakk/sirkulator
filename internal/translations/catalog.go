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
	"%d hits (%v)":                          40,
	"1 per line":                            16,
	"About":                                 88,
	"Actions":                               58,
	"Add new schedule":                      96,
	"Agent":                                 84,
	"Already in catalogue":                  33,
	"Are you sure?":                         86,
	"Associated country/area":               64,
	"Associated nationality":                65,
	"Audience":                              78,
	"Basic information":                     41,
	"Binding":                               80,
	"Birthyear":                             66,
	"Broader terms":                         26,
	"Cancel":                                59,
	"Choose job":                            98,
	"Circulation":                           1,
	"Configuration":                         5,
	"Content":                               72,
	"Contributions and relations":           38,
	"Cover-image":                           36,
	"Created":                               34,
	"Cron expression":                       99,
	"Data":                                  95,
	"Deathyear":                             67,
	"Delete":                                87,
	"Description (short)":                   62,
	"Dewey number":                          55,
	"Dewey numbers where %s is a component": 30,
	"Discontinued":                          92,
	"Disestablishment year":                 51,
	"Established":                           91,
	"Fiction":                               75,
	"Foundation year":                       49,
	"Gender":                                63,
	"Genre and forms":                       77,
	"Has components":                        28,
	"Holdings":                              4,
	"Home":                                  0,
	"ISBN, ISSN or EAN":                     15,
	"Identificators and links":              20,
	"Identifiers":                           14,
	"Import":                                13,
	"Job":                                   97,
	"Latest job runs":                       8,
	"Lifespan":                              48,
	"Main language":                         73,
	"Maintenance":                           6,
	"Metadata":                              3,
	"Must be an integer":                    82,
	"Name":                                  42,
	"Name variations":                       45,
	"Narrower terms":                        27,
	"Next page":                             54,
	"Nonfiction":                            76,
	"Notes":                                 89,
	"Number of pages":                       81,
	"One entry per line":                    46,
	"Orders":                                2,
	"Other languages":                       74,
	"Other relations":                       25,
	"Parent name":                           47,
	"Personalia":                            61,
	"Physical characteristics":              79,
	"Preview":                               17,
	"Previous page":                         53,
	"Properties":                            19,
	"Publication":                           24,
	"Publication cover-image":               37,
	"Publications":                          39,
	"Publications and contributions":        21,
	"Publications classified with":          31,
	"Publisher":                             70,
	"Reference terms":                       29,
	"Relation":                              94,
	"Required field":                        43,
	"Resource":                              93,
	"Role":                                  22,
	"Role/relation":                         83,
	"Run now (one-off)":                     101,
	"Schedule job":                          100,
	"Scheduled jobs":                        9,
	"Schedules":                             102,
	"Search and connect to resource":        85,
	"Search/browse catalogue":               11,
	"Short description":                     44,
	"Show metadata for review":              10,
	"Show recent transactions":              7,
	"Started (duration)":                    56,
	"Status":                                57,
	"Subtitle":                              69,
	"Title":                                 68,
	"Uncertain":                             52,
	"Updated":                               35,
	"View output":                           60,
	"Year":                                  23,
	"Year must be a 1-4 digit number. Negative numbers signify BCE.": 50,
	"Year must be a 4-digit number":                                  71,
	"Years of activity":                                              90,
	"include archived":                                               12,
	"include narrower numbers":                                       32,
	"wait...":                                                        18,
}

var enIndex = []uint32{ // 104 elements
	// Entry 0 - 1F
	0x00000000, 0x00000005, 0x00000011, 0x00000018,
	0x00000021, 0x0000002a, 0x00000038, 0x00000044,
	0x0000005d, 0x0000006d, 0x0000007c, 0x00000095,
	0x000000ad, 0x000000be, 0x000000c5, 0x000000d1,
	0x000000e3, 0x000000ee, 0x000000f6, 0x000000fe,
	0x00000109, 0x00000122, 0x00000141, 0x00000146,
	0x0000014b, 0x00000157, 0x00000167, 0x00000175,
	0x00000184, 0x00000193, 0x000001a3, 0x000001cc,
	// Entry 20 - 3F
	0x000001e9, 0x00000202, 0x00000217, 0x0000021f,
	0x00000227, 0x00000233, 0x0000024b, 0x00000267,
	0x00000274, 0x00000287, 0x00000299, 0x0000029e,
	0x000002ad, 0x000002bf, 0x000002cf, 0x000002e2,
	0x000002ee, 0x000002f7, 0x00000307, 0x00000346,
	0x0000035c, 0x00000366, 0x00000374, 0x0000037e,
	0x0000038b, 0x0000039e, 0x000003a5, 0x000003ad,
	0x000003b4, 0x000003c0, 0x000003cb, 0x000003df,
	// Entry 40 - 5F
	0x000003e6, 0x000003fe, 0x00000415, 0x0000041f,
	0x00000429, 0x0000042f, 0x00000438, 0x00000442,
	0x00000460, 0x00000468, 0x00000476, 0x00000486,
	0x0000048e, 0x00000499, 0x000004a9, 0x000004b2,
	0x000004cb, 0x000004d3, 0x000004e3, 0x000004f6,
	0x00000504, 0x0000050a, 0x00000529, 0x00000537,
	0x0000053e, 0x00000544, 0x0000054a, 0x0000055c,
	0x00000568, 0x00000575, 0x0000057e, 0x00000587,
	// Entry 60 - 7F
	0x0000058c, 0x0000059d, 0x000005a1, 0x000005ac,
	0x000005bc, 0x000005c9, 0x000005db, 0x000005e5,
} // Size: 440 bytes

const enData string = "" + // Size: 1509 bytes
	"\x02Home\x02Circulation\x02Orders\x02Metadata\x02Holdings\x02Configurati" +
	"on\x02Maintenance\x02Show recent transactions\x02Latest job runs\x02Sche" +
	"duled jobs\x02Show metadata for review\x02Search/browse catalogue\x02inc" +
	"lude archived\x02Import\x02Identifiers\x02ISBN, ISSN or EAN\x021 per lin" +
	"e\x02Preview\x02wait...\x02Properties\x02Identificators and links\x02Pub" +
	"lications and contributions\x02Role\x02Year\x02Publication\x02Other rela" +
	"tions\x02Broader terms\x02Narrower terms\x02Has components\x02Reference " +
	"terms\x02Dewey numbers where %[1]s is a component\x02Publications classi" +
	"fied with\x02include narrower numbers\x02Already in catalogue\x02Created" +
	"\x02Updated\x02Cover-image\x02Publication cover-image\x02Contributions a" +
	"nd relations\x02Publications\x02%[1]d hits (%[2]v)\x02Basic information" +
	"\x02Name\x02Required field\x02Short description\x02Name variations\x02On" +
	"e entry per line\x02Parent name\x02Lifespan\x02Foundation year\x02Year m" +
	"ust be a 1-4 digit number. Negative numbers signify BCE.\x02Disestablish" +
	"ment year\x02Uncertain\x02Previous page\x02Next page\x02Dewey number\x02" +
	"Started (duration)\x02Status\x02Actions\x02Cancel\x02View output\x02Pers" +
	"onalia\x02Description (short)\x02Gender\x02Associated country/area\x02As" +
	"sociated nationality\x02Birthyear\x02Deathyear\x02Title\x02Subtitle\x02P" +
	"ublisher\x02Year must be a 4-digit number\x02Content\x02Main language" +
	"\x02Other languages\x02Fiction\x02Nonfiction\x02Genre and forms\x02Audie" +
	"nce\x02Physical characteristics\x02Binding\x02Number of pages\x02Must be" +
	" an integer\x02Role/relation\x02Agent\x02Search and connect to resource" +
	"\x02Are you sure?\x02Delete\x02About\x02Notes\x02Years of activity\x02Es" +
	"tablished\x02Discontinued\x02Resource\x02Relation\x02Data\x02Add new sch" +
	"edule\x02Job\x02Choose job\x02Cron expression\x02Schedule job\x02Run now" +
	" (one-off)\x02Schedules"

var noIndex = []uint32{ // 104 elements
	// Entry 0 - 1F
	0x00000000, 0x00000005, 0x00000011, 0x0000001e,
	0x00000027, 0x0000002f, 0x0000003d, 0x00000049,
	0x00000061, 0x00000072, 0x00000087, 0x000000a7,
	0x000000bc, 0x000000cf, 0x000000d8, 0x000000e8,
	0x000000fd, 0x00000109, 0x00000116, 0x0000011e,
	0x00000129, 0x00000143, 0x00000158, 0x0000015e,
	0x00000162, 0x0000016c, 0x0000017d, 0x00000190,
	0x000001a4, 0x000001b3, 0x000001c5, 0x000001e4,
	// Entry 20 - 3F
	0x00000200, 0x0000021c, 0x00000231, 0x0000023b,
	0x00000242, 0x0000024f, 0x00000268, 0x0000027d,
	0x00000288, 0x0000029c, 0x000002b6, 0x000002bb,
	0x000002c9, 0x000002da, 0x000002e9, 0x00000301,
	0x00000315, 0x0000031d, 0x00000326, 0x00000364,
	0x0000036d, 0x00000376, 0x00000383, 0x0000038e,
	0x0000039a, 0x000003ad, 0x000003b4, 0x000003bf,
	0x000003c6, 0x000003d1, 0x000003dc, 0x000003ed,
	// Entry 40 - 5F
	0x000003f4, 0x0000040b, 0x00000422, 0x0000042e,
	0x00000437, 0x0000043e, 0x0000044a, 0x00000452,
	0x00000475, 0x0000047d, 0x0000048d, 0x0000049a,
	0x000004a2, 0x000004a6, 0x000004b6, 0x000004c1,
	0x000004d4, 0x000004e0, 0x000004e9, 0x000004fe,
	0x0000050d, 0x00000514, 0x0000052e, 0x0000053c,
	0x00000542, 0x00000545, 0x0000054b, 0x0000055f,
	0x00000569, 0x00000572, 0x0000057a, 0x00000583,
	// Entry 60 - 7F
	0x00000588, 0x0000059d, 0x000005a2, 0x000005ac,
	0x000005b9, 0x000005c2, 0x000005d6, 0x000005eb,
} // Size: 440 bytes

const noData string = "" + // Size: 1515 bytes
	"\x02Hjem\x02Sirkulasjon\x02Bestillinger\x02Metadata\x02Bestand\x02Konfig" +
	"urasjon\x02Vedlikehold\x02Vis siste transaksjoner\x02Siste kjøringer\x02" +
	"Planlagte kjøringer\x02Vis opplysninger til gjennomsyn\x02Søk/bla i kata" +
	"logen\x02inkluder arkiverte\x02Importer\x02Identifikatorer\x02ISBN, ISSN" +
	" eller EAN\x021 per linje\x02Forhåndsvis\x02vent...\x02Egenskaper\x02Ide" +
	"ntifikatorer og lenker\x02Utgivelser og bidrag\x02Rolle\x02År\x02Utgivel" +
	"se\x02Andre relasjoner\x02Overordnede begrep\x02Underordnede begrep\x02E" +
	"r oppbyggd av\x02Henvisningstermer\x02Deweynummer hvor %[1]s inngår\x02U" +
	"tgivelser klassifisert med\x02inkluder underordnede numre\x02Allerede i " +
	"katalogen\x02Opprettet\x02Endret\x02Forsidebilde\x02Utgivelsens forsideb" +
	"ilde\x02Bidrag og relasjoner\x02Utgivelser\x02%[1]d treff (%[2]v)\x02Gru" +
	"nnleggende informasjon\x02Navn\x02Påkrevd felt\x02Kort beskrivelse\x02Na" +
	"vnevarianter\x02En innførsel per linje\x02Navn på overordnet\x02Levetid" +
	"\x02Etablert\x02År må være et 1-4 sifret heltall. Negative tall betyr BC" +
	"E.\x02Oppløst\x02Usikkert\x02Forrige side\x02Neste side\x02Deweynummer" +
	"\x02Startet (varighet)\x02Status\x02Handlinger\x02Avbryt\x02Vis utdata" +
	"\x02Personalia\x02Kort beskrivelse\x02Kjønn\x02Assosiert land/område\x02" +
	"Assosiert nasjonalitet\x02Fødselsår\x02Dødsår\x02Tittel\x02Undertittel" +
	"\x02Utgiver\x02År må være et 4-siffret heltall\x02Innhold\x02Språk (hove" +
	"d-)\x02Andre språk\x02Fiksjon\x02Fag\x02Sjanger og form\x02Målgruppe\x02" +
	"Fysiske egenskaper\x02Innbdinding\x02Sidetall\x02Må være et heltall\x02R" +
	"olle/relasjon\x02Aktør\x02Søk og koble til ressurs\x02Er du sikker?\x02S" +
	"lett\x02Om\x02Noter\x02Aktiv fra-til (år)\x02Grunnlagt\x02Opphørt\x02Res" +
	"surs\x02Relasjon\x02Data\x02Sett opp ny kjøring\x02Jobb\x02Velg jobb\x02" +
	"Cron-uttrykk\x02Legg til\x02Kjør nå (en gang)\x02Planlagte kjøringer"

	// Total table size 3904 bytes (3KiB); checksum: 9A80CA12
