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
	"%d hits (%v)":                          38,
	"1 per line":                            16,
	"About":                                 86,
	"Actions":                               57,
	"Add new schedule":                      94,
	"Agent":                                 82,
	"Already in catalogue":                  33,
	"Archived":                              106,
	"Are you sure?":                         84,
	"Associated country/area":               63,
	"Associated nationality":                64,
	"Audience":                              76,
	"Basic information":                     39,
	"Binding":                               78,
	"Birthyear":                             65,
	"Broader terms":                         26,
	"Cancel":                                58,
	"Choose job":                            96,
	"Circulation":                           1,
	"Configuration":                         5,
	"Content":                               70,
	"Contributions and relations":           36,
	"Cover-image":                           34,
	"Created":                               104,
	"Cron expression":                       97,
	"Data":                                  93,
	"Deathyear":                             66,
	"Delete":                                85,
	"Description (short)":                   61,
	"Dewey number":                          53,
	"Dewey numbers where %s is a component": 30,
	"Discontinued":                          90,
	"Disestablishment year":                 49,
	"Established":                           89,
	"Fiction":                               73,
	"Foundation year":                       47,
	"Gender":                                62,
	"Genre and forms":                       75,
	"Has components":                        28,
	"Holdings":                              4,
	"Home":                                  0,
	"ISBN, ISSN or EAN":                     15,
	"Identificators and links":              54,
	"Identifiers":                           14,
	"Import":                                13,
	"Job":                                   95,
	"Latest job runs":                       8,
	"Lifespan":                              46,
	"Local and external descriptions":       20,
	"Main language":                         71,
	"Maintenance":                           6,
	"Metadata":                              3,
	"Must be an integer":                    80,
	"Name":                                  40,
	"Name variations":                       43,
	"Narrower terms":                        27,
	"Next page":                             52,
	"Nonfiction":                            74,
	"Notes":                                 87,
	"Number of pages":                       79,
	"One entry per line":                    44,
	"Orders":                                2,
	"Other languages":                       72,
	"Other relations":                       25,
	"Parent name":                           45,
	"Personalia":                            60,
	"Physical characteristics":              77,
	"Preview":                               17,
	"Previous page":                         51,
	"Properties":                            19,
	"Publication":                           24,
	"Publication cover-image":               35,
	"Publications":                          37,
	"Publications and contributions":        21,
	"Publications classified with":          31,
	"Reference terms":                       29,
	"Relation":                              92,
	"Required field":                        41,
	"Resource":                              91,
	"Role":                                  22,
	"Role/relation":                         81,
	"Run now (one-off)":                     99,
	"Schedule job":                          98,
	"Scheduled jobs":                        9,
	"Schedules":                             100,
	"Search and connect to resource":        83,
	"Search/browse catalogue":               11,
	"Short description":                     42,
	"Show metadata for review":              10,
	"Show recent transactions":              7,
	"Started (duration)":                    55,
	"Status":                                56,
	"Subtitle":                              68,
	"This resource is archived":             102,
	"Title":                                 67,
	"Uncertain":                             50,
	"Updated":                               105,
	"View output":                           59,
	"Year":                                  23,
	"Year must be a 1-4 digit number. Negative numbers signify BCE.": 48,
	"Year must be a 4-digit number":                                  69,
	"Years of activity":                                              88,
	"include archived":                                               12,
	"include narrower numbers":                                       32,
	"restore":                                                        103,
	"save":                                                           101,
	"wait...":                                                        18,
}

var enIndex = []uint32{ // 108 elements
	// Entry 0 - 1F
	0x00000000, 0x00000005, 0x00000011, 0x00000018,
	0x00000021, 0x0000002a, 0x00000038, 0x00000044,
	0x0000005d, 0x0000006d, 0x0000007c, 0x00000095,
	0x000000ad, 0x000000be, 0x000000c5, 0x000000d1,
	0x000000e3, 0x000000ee, 0x000000f6, 0x000000fe,
	0x00000109, 0x00000129, 0x00000148, 0x0000014d,
	0x00000152, 0x0000015e, 0x0000016e, 0x0000017c,
	0x0000018b, 0x0000019a, 0x000001aa, 0x000001d3,
	// Entry 20 - 3F
	0x000001f0, 0x00000209, 0x0000021e, 0x0000022a,
	0x00000242, 0x0000025e, 0x0000026b, 0x0000027e,
	0x00000290, 0x00000295, 0x000002a4, 0x000002b6,
	0x000002c6, 0x000002d9, 0x000002e5, 0x000002ee,
	0x000002fe, 0x0000033d, 0x00000353, 0x0000035d,
	0x0000036b, 0x00000375, 0x00000382, 0x0000039b,
	0x000003ae, 0x000003b5, 0x000003bd, 0x000003c4,
	0x000003d0, 0x000003db, 0x000003ef, 0x000003f6,
	// Entry 40 - 5F
	0x0000040e, 0x00000425, 0x0000042f, 0x00000439,
	0x0000043f, 0x00000448, 0x00000466, 0x0000046e,
	0x0000047c, 0x0000048c, 0x00000494, 0x0000049f,
	0x000004af, 0x000004b8, 0x000004d1, 0x000004d9,
	0x000004e9, 0x000004fc, 0x0000050a, 0x00000510,
	0x0000052f, 0x0000053d, 0x00000544, 0x0000054a,
	0x00000550, 0x00000562, 0x0000056e, 0x0000057b,
	0x00000584, 0x0000058d, 0x00000592, 0x000005a3,
	// Entry 60 - 7F
	0x000005a7, 0x000005b2, 0x000005c2, 0x000005cf,
	0x000005e1, 0x000005eb, 0x000005f0, 0x0000060a,
	0x00000612, 0x0000061a, 0x00000622, 0x0000062b,
} // Size: 456 bytes

const enData string = "" + // Size: 1579 bytes
	"\x02Home\x02Circulation\x02Orders\x02Metadata\x02Holdings\x02Configurati" +
	"on\x02Maintenance\x02Show recent transactions\x02Latest job runs\x02Sche" +
	"duled jobs\x02Show metadata for review\x02Search/browse catalogue\x02inc" +
	"lude archived\x02Import\x02Identifiers\x02ISBN, ISSN or EAN\x021 per lin" +
	"e\x02Preview\x02wait...\x02Properties\x02Local and external descriptions" +
	"\x02Publications and contributions\x02Role\x02Year\x02Publication\x02Oth" +
	"er relations\x02Broader terms\x02Narrower terms\x02Has components\x02Ref" +
	"erence terms\x02Dewey numbers where %[1]s is a component\x02Publications" +
	" classified with\x02include narrower numbers\x02Already in catalogue\x02" +
	"Cover-image\x02Publication cover-image\x02Contributions and relations" +
	"\x02Publications\x02%[1]d hits (%[2]v)\x02Basic information\x02Name\x02R" +
	"equired field\x02Short description\x02Name variations\x02One entry per l" +
	"ine\x02Parent name\x02Lifespan\x02Foundation year\x02Year must be a 1-4 " +
	"digit number. Negative numbers signify BCE.\x02Disestablishment year\x02" +
	"Uncertain\x02Previous page\x02Next page\x02Dewey number\x02Identificator" +
	"s and links\x02Started (duration)\x02Status\x02Actions\x02Cancel\x02View" +
	" output\x02Personalia\x02Description (short)\x02Gender\x02Associated cou" +
	"ntry/area\x02Associated nationality\x02Birthyear\x02Deathyear\x02Title" +
	"\x02Subtitle\x02Year must be a 4-digit number\x02Content\x02Main languag" +
	"e\x02Other languages\x02Fiction\x02Nonfiction\x02Genre and forms\x02Audi" +
	"ence\x02Physical characteristics\x02Binding\x02Number of pages\x02Must b" +
	"e an integer\x02Role/relation\x02Agent\x02Search and connect to resource" +
	"\x02Are you sure?\x02Delete\x02About\x02Notes\x02Years of activity\x02Es" +
	"tablished\x02Discontinued\x02Resource\x02Relation\x02Data\x02Add new sch" +
	"edule\x02Job\x02Choose job\x02Cron expression\x02Schedule job\x02Run now" +
	" (one-off)\x02Schedules\x02save\x02This resource is archived\x02restore" +
	"\x02Created\x02Updated\x02Archived"

var noIndex = []uint32{ // 108 elements
	// Entry 0 - 1F
	0x00000000, 0x00000005, 0x00000011, 0x0000001e,
	0x00000027, 0x0000002f, 0x0000003d, 0x00000049,
	0x00000061, 0x00000072, 0x00000087, 0x000000a7,
	0x000000bc, 0x000000cf, 0x000000d8, 0x000000e8,
	0x000000fd, 0x00000109, 0x00000116, 0x0000011e,
	0x00000129, 0x0000014a, 0x0000015f, 0x00000165,
	0x00000169, 0x00000173, 0x00000184, 0x00000197,
	0x000001ab, 0x000001ba, 0x000001cc, 0x000001eb,
	// Entry 20 - 3F
	0x00000207, 0x00000223, 0x00000238, 0x00000245,
	0x0000025e, 0x00000273, 0x0000027e, 0x00000292,
	0x000002ac, 0x000002b1, 0x000002bf, 0x000002d0,
	0x000002df, 0x000002f7, 0x0000030b, 0x00000313,
	0x0000031c, 0x0000035a, 0x00000363, 0x0000036c,
	0x00000379, 0x00000384, 0x00000390, 0x000003aa,
	0x000003bd, 0x000003c4, 0x000003cf, 0x000003d6,
	0x000003e1, 0x000003ec, 0x000003fd, 0x00000404,
	// Entry 40 - 5F
	0x0000041b, 0x00000432, 0x0000043e, 0x00000447,
	0x0000044e, 0x0000045a, 0x0000047d, 0x00000485,
	0x00000495, 0x000004a2, 0x000004aa, 0x000004ae,
	0x000004be, 0x000004c9, 0x000004dc, 0x000004e8,
	0x000004f1, 0x00000506, 0x00000515, 0x0000051c,
	0x00000536, 0x00000544, 0x0000054a, 0x0000054d,
	0x00000553, 0x00000567, 0x00000571, 0x0000057a,
	0x00000582, 0x0000058b, 0x00000590, 0x000005a5,
	// Entry 60 - 7F
	0x000005aa, 0x000005b4, 0x000005c1, 0x000005ca,
	0x000005de, 0x000005f3, 0x000005f9, 0x00000615,
	0x00000621, 0x0000062b, 0x00000632, 0x0000063b,
} // Size: 456 bytes

const noData string = "" + // Size: 1595 bytes
	"\x02Hjem\x02Sirkulasjon\x02Bestillinger\x02Metadata\x02Bestand\x02Konfig" +
	"urasjon\x02Vedlikehold\x02Vis siste transaksjoner\x02Siste kj??ringer\x02" +
	"Planlagte kj??ringer\x02Vis opplysninger til gjennomsyn\x02S??k/bla i kata" +
	"logen\x02inkluder arkiverte\x02Importer\x02Identifikatorer\x02ISBN, ISSN" +
	" eller EAN\x021 per linje\x02Forh??ndsvis\x02vent...\x02Egenskaper\x02Int" +
	"erne og eksterne beskrivelser\x02Utgivelser og bidrag\x02Rolle\x02??r\x02" +
	"Utgivelse\x02Andre relasjoner\x02Overordnede begrep\x02Underordnede begr" +
	"ep\x02Er oppbyggd av\x02Henvisningstermer\x02Deweynummer hvor %[1]s inng" +
	"??r\x02Utgivelser klassifisert med\x02inkluder underordnede numre\x02All" +
	"erede i katalogen\x02Forsidebilde\x02Utgivelsens forsidebilde\x02Bidrag " +
	"og relasjoner\x02Utgivelser\x02%[1]d treff (%[2]v)\x02Grunnleggende info" +
	"rmasjon\x02Navn\x02P??krevd felt\x02Kort beskrivelse\x02Navnevarianter" +
	"\x02En innf??rsel per linje\x02Navn p?? overordnet\x02Levetid\x02Etablert" +
	"\x02??r m?? v??re et 1-4 sifret heltall. Negative tall betyr BCE.\x02Oppl??s" +
	"t\x02Usikkert\x02Forrige side\x02Neste side\x02Deweynummer\x02Identifika" +
	"torer og lenker\x02Startet (varighet)\x02Status\x02Handlinger\x02Avbryt" +
	"\x02Vis utdata\x02Personalia\x02Kort beskrivelse\x02Kj??nn\x02Assosiert l" +
	"and/omr??de\x02Assosiert nasjonalitet\x02F??dsels??r\x02D??ds??r\x02Tittel" +
	"\x02Undertittel\x02??r m?? v??re et 4-siffret heltall\x02Innhold\x02Spr??k (" +
	"hoved-)\x02Andre spr??k\x02Fiksjon\x02Fag\x02Sjanger og form\x02M??lgruppe" +
	"\x02Fysiske egenskaper\x02Innbdinding\x02Sidetall\x02M?? v??re et heltall" +
	"\x02Rolle/relasjon\x02Akt??r\x02S??k og koble til ressurs\x02Er du sikker?" +
	"\x02Slett\x02Om\x02Noter\x02Aktiv fra-til (??r)\x02Grunnlagt\x02Opph??rt" +
	"\x02Ressurs\x02Relasjon\x02Data\x02Sett opp ny kj??ring\x02Jobb\x02Velg j" +
	"obb\x02Cron-uttrykk\x02Legg til\x02Kj??r n?? (en gang)\x02Planlagte kj??rin" +
	"ger\x02lagre\x02Denne ressursen er akrivert\x02gjenopprett\x02Opprettet" +
	"\x02Endret\x02Arkivert"

	// Total table size 4086 bytes (3KiB); checksum: F4861F79
