package etl

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/marc"
)

func TestParseYearRange(t *testing.T) {
	tests := []struct {
		input string
		want  sirkulator.YearRange
	}{
		{"1981", sirkulator.YearRange{FromYear: 1981}},
		{"  1981", sirkulator.YearRange{FromYear: 1981}},
		{"1949-", sirkulator.YearRange{FromYear: 1949}},
		{"1828-1906", sirkulator.YearRange{FromYear: 1828, ToYear: 1906}},
		{"1828 - 1906", sirkulator.YearRange{FromYear: 1828, ToYear: 1906}},
	}

	for _, test := range tests {
		if got := parseYearRange(test.input); got != test.want {
			t.Errorf("yearRange(%q): got %v; want %v", test.input, got, test.want)
		}
	}
}

func TestParsePages(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"491", 491},
		{" 491", 491},
		{"491 sider", 491},
		{"12 s", 12},
	}

	for _, test := range tests {
		if got := parsePages(test.input); got != test.want {
			t.Errorf("parsePages(%q): got %v; want %v", test.input, got, test.want)
		}
	}
}

/*func TestPublicationLabel(t *testing.T) {
	tests := []struct {
		pub       sirkulator.Publication
		agents []sirkulator.Agent
		relations []sirkulator.Relation
		want      string
	}{
		{
			pub: sirkulator.Publication{
				Title: "hei",
			},
			want: "hei",
		},
		{
			pub: sirkulator.Publication{
				Title: "hei",
			},
			relations: []sirkulator.Relation{{Type: "has_contributor"}},
			want:      "hei",
		},
	}

	for _, test := range tests {
		if got := publicationLabel(test.pub, test.relations); got != test.want {
			t.Errorf("publicationLabel(%v %v): got %v; want %v", test.pub, test.relations, got, test.want)
		}
	}
}*/

func testID() func() string {
	i := 0
	return func() string {
		i++
		return fmt.Sprintf("t%d", i)
	}
}

func TestIngestOAIRecord(t *testing.T) {
	t.Run("first edition monograph", func(t *testing.T) {
		const isbn9788203365133 = `
			<record xmlns="http://www.loc.gov/MARC21/slim" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.loc.gov/MARC21/slim http://www.loc.gov/standards/marcxml/schema/MARC21slim.xsd">
				<leader>02589cam a2200397 c 4500</leader>
				<controlfield tag="005">20211216084534.0</controlfield>
				<controlfield tag="007">ta</controlfield>
				<controlfield tag="008">210629s2021    no     e ||||||0| 0dnob|</controlfield>
				<controlfield tag="001">999921380896302201</controlfield>
				<datafield tag="020" ind1=" " ind2=" ">
					<subfield code="a">9788203365133</subfield>
					<subfield code="c">Nkr 429.00</subfield>
					<subfield code="q">innbundet</subfield>
				</datafield>
				<datafield tag="035" ind1=" " ind2=" ">
					<subfield code="a">(NO-OsBA)0621963</subfield>
				</datafield>
				<datafield tag="035" ind1=" " ind2=" ">
					<subfield code="a">oai:bibbi.bs.no:0621963</subfield>
				</datafield>
				<datafield tag="040" ind1=" " ind2=" ">
					<subfield code="a">NO-OsBA</subfield>
					<subfield code="b">nob</subfield>
					<subfield code="e">rda</subfield>
					<subfield code="d">NO-OsNB</subfield>
				</datafield>
				<datafield tag="082" ind1="0" ind2="4">
					<subfield code="a">839.82374</subfield>
					<subfield code="q">NO-OsBA</subfield>
					<subfield code="2">23/nor</subfield>
				</datafield>
				<datafield tag="100" ind1="1" ind2=" ">
					<subfield code="a">Køltzow, Liv</subfield>
					<subfield code="d">1945-</subfield>
					<subfield code="0">(NO-TrBIB)90086277</subfield>
					<subfield code="4">aut</subfield>
				</datafield>
				<datafield tag="240" ind1="1" ind2="0">
					<subfield code="a">Liv Køltzow</subfield>
				</datafield>
				<datafield tag="245" ind1="1" ind2="0">
					<subfield code="a">Liv Køltzow :</subfield>
					<subfield code="b">dagbøker i utvalg 1964-2008</subfield>
					<subfield code="c">Hans Petter Blad og Kaja Schjerven Mollerin (red.)</subfield>
				</datafield>
				<datafield tag="264" ind1=" " ind2="1">
					<subfield code="a">Oslo</subfield>
					<subfield code="b">Aschehoug</subfield>
					<subfield code="c">2021</subfield>
				</datafield>
				<datafield tag="300" ind1=" " ind2=" ">
					<subfield code="a">491 sider</subfield>
					<subfield code="c">21 cm</subfield>
				</datafield>
				<datafield tag="336" ind1=" " ind2=" ">
					<subfield code="a">tekst</subfield>
					<subfield code="0">http://rdaregistry.info/termList/RDAContentType/1020</subfield>
					<subfield code="2">rdaco</subfield>
				</datafield>
				<datafield tag="337" ind1=" " ind2=" ">
					<subfield code="a">uformidlet</subfield>
					<subfield code="0">http://rdaregistry.info/termList/RDAMediaType/1007</subfield>
					<subfield code="2">rdamt</subfield>
				</datafield>
				<datafield tag="338" ind1=" " ind2=" ">
					<subfield code="a">bind</subfield>
					<subfield code="0">http://rdaregistry.info/termList/RDACarrierType/1049</subfield>
					<subfield code="2">rdact</subfield>
				</datafield>
				<datafield tag="520" ind1=" " ind2=" ">
					<subfield code="a">Siden 1964 har forfatter Liv Køltzow ført dagbok. I notatbøker og spiralhefter har hun skrevet om seg selv i livet, i fortid, nåtid og fremtid, og om menneskene som har hatt plass i livet hennes. Disse tankene flettes sammen med refleksjoner om litteratur og hennes egen skriveprosess. Nå foreligger et utvalg av disse 58 notatbøker. Et tverrsnitt av levd liv, et blikk på et sentralt forfatterskap i norsk etterkrigstid og et litteraturhistorisk dokument. Omtalen er utarbeidet av BS.</subfield>
				</datafield>
				<datafield tag="655" ind1=" " ind2="7">
					<subfield code="a">Dagbøker</subfield>
					<subfield code="0">https://id.nb.no/vocabulary/ntsf/54</subfield>
					<subfield code="2">ntsf</subfield>
					<subfield code="9">nob</subfield>
				</datafield>
				<datafield tag="655" ind1=" " ind2="7">
					<subfield code="a">Dagbøker</subfield>
					<subfield code="0">https://id.nb.no/vocabulary/ntsf/54</subfield>
					<subfield code="2">ntsf</subfield>
					<subfield code="9">nno</subfield>
				</datafield>
				<datafield tag="653" ind1=" " ind2="0">
					<subfield code="a">dagbøker</subfield>
				</datafield>
				<datafield tag="600" ind1="1" ind2="7">
					<subfield code="a">Køltzow, Liv</subfield>
					<subfield code="d">1945-</subfield>
					<subfield code="0">(NO-TrBIB)90086277</subfield>
					<subfield code="2">bare</subfield>
				</datafield>
				<datafield tag="700" ind1="1" ind2=" ">
					<subfield code="a">Blad, Hans Petter</subfield>
					<subfield code="d">1962-</subfield>
					<subfield code="4">edt</subfield>
					<subfield code="0">(NO-TrBIB)90916002</subfield>
				</datafield>
				<datafield tag="700" ind1="1" ind2=" ">
					<subfield code="a">Mollerin, Kaja Schjerven</subfield>
					<subfield code="d">1980-</subfield>
					<subfield code="4">edt</subfield>
					<subfield code="0">(NO-TrBIB)6088516</subfield>
				</datafield>
				<datafield tag="856" ind1="4" ind2="2">
					<subfield code="3">Miniatyrbilde</subfield>
					<subfield code="u">https://contents.bibs.aws.unit.no/files/images/small/3/3/9788203365133.jpg</subfield>
					<subfield code="q">image/jpeg</subfield>
				</datafield>
				<datafield tag="856" ind1="4" ind2="2">
					<subfield code="3">Omslagsbilde</subfield>
					<subfield code="u">https://contents.bibs.aws.unit.no/files/images/large/3/3/9788203365133.jpg</subfield>
					<subfield code="q">image/jpeg</subfield>
				</datafield>
				<datafield tag="856" ind1="4" ind2="2">
					<subfield code="3">Originalt bilde</subfield>
					<subfield code="u">https://contents.bibs.aws.unit.no/files/images/original/3/3/9788203365133.jpg</subfield>
					<subfield code="q">image/jpeg</subfield>
				</datafield>
				<datafield tag="856" ind1="4" ind2="2">
					<subfield code="3">Forlagets beskrivelse (lang)</subfield>
					<subfield code="u">https://contents.bibs.aws.unit.no/content/?isbn=9788203365133</subfield>
				</datafield>
				<datafield tag="856" ind1=" " ind2=" ">
					<subfield code="a">aja.bs.no</subfield>
					<subfield code="n">Biblioteksentralen, Oslo</subfield>
					<subfield code="q">image/jpeg</subfield>
					<subfield code="u">https://media.aja.bs.no/4bfe1c40-8ee2-4601-b7eb-2e032a2e59b7/cover/original.jpg</subfield>
					<subfield code="3">Omslagsbilde</subfield>
				</datafield>
				<datafield tag="856" ind1=" " ind2=" ">
					<subfield code="a">aja.bs.no</subfield>
					<subfield code="n">Biblioteksentralen, Oslo</subfield>
					<subfield code="q">image/jpeg</subfield>
					<subfield code="u">https://media.aja.bs.no/4bfe1c40-8ee2-4601-b7eb-2e032a2e59b7/cover/thumbnail.jpg</subfield>
					<subfield code="3">Miniatyrbilde</subfield>
				</datafield>
				<datafield tag="913" ind1=" " ind2=" ">
					<subfield code="a">Norbok</subfield>
					<subfield code="b">NB</subfield>
				</datafield>
			</record>
			`
		want := Ingestion{
			Resources: []sirkulator.Resource{
				{
					ID:    "t1",
					Type:  sirkulator.TypePublication,
					Label: "Liv Køltzow - Liv Køltzow: dagbøker i utvalg 1964-2008 (2021)",
					Data: sirkulator.Publication{
						Title:      "Liv Køltzow",
						Subtitle:   "dagbøker i utvalg 1964-2008",
						Language:   "nob",
						GenreForms: []string{"Dagbøker"},
						Nonfiction: true,
						Year:       2021,
						//YearFirst:  2021, TODO later
						Publisher: "Aschehoug",
						NumPages:  491,
					},
				},
				{
					ID:    "t2",
					Type:  sirkulator.TypePerson,
					Label: "Liv Køltzow (1945-)",
					Links: [][2]string{{"bibsys", "90086277"}},
					Data: sirkulator.Person{
						Name: "Liv Køltzow",
						YearRange: sirkulator.YearRange{
							FromYear: 1945,
						},
					},
				},
				{
					ID:    "t3",
					Type:  sirkulator.TypePerson,
					Label: "Hans Petter Blad (1962-)",
					Links: [][2]string{{"bibsys", "90916002"}},
					Data: sirkulator.Person{
						Name: "Hans Petter Blad",
						YearRange: sirkulator.YearRange{
							FromYear: 1962,
						},
					},
				},
				{
					ID:    "t4",
					Type:  sirkulator.TypePerson,
					Label: "Kaja Schjerven Mollerin (1980-)",
					Links: [][2]string{{"bibsys", "6088516"}},
					Data: sirkulator.Person{
						Name: "Kaja Schjerven Mollerin",
						YearRange: sirkulator.YearRange{
							FromYear: 1980,
						},
					},
				},
			},
			Relations: []sirkulator.Relation{
				{
					FromID: "t1",
					ToID:   "t2",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "aut", "main_entry": true},
				},
				{
					FromID: "t1",
					ToID:   "t2",
					Type:   "has_subject",
				},
				{
					FromID: "t1",
					ToID:   "t3",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "edt"},
				},
				{
					FromID: "t1",
					ToID:   "t4",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "edt"},
				},
			},
			Covers: []FileFetch{
				{
					ResourceID: "t1",
					URL:        "https://contents.bibs.aws.unit.no/files/images/original/3/3/9788203365133.jpg",
				},
				{
					ResourceID: "t1",
					URL:        "https://media.aja.bs.no/4bfe1c40-8ee2-4601-b7eb-2e032a2e59b7/cover/original.jpg",
				},
				{
					ResourceID: "t1",
					URL:        "https://contents.bibs.aws.unit.no/files/images/small/3/3/9788203365133.jpg",
				},
				{
					ResourceID: "t1",
					URL:        "https://contents.bibs.aws.unit.no/files/images/large/3/3/9788203365133.jpg",
				},
				{
					ResourceID: "t1",
					URL:        "https://media.aja.bs.no/4bfe1c40-8ee2-4601-b7eb-2e032a2e59b7/cover/thumbnail.jpg",
				},
			},
			Reviews: []sirkulator.Relation{
				{
					FromID: "t1",
					Type:   "published_by",
					Data:   map[string]interface{}{"label": "Aschehoug"},
				},
			},
		}

		got, err := ingestMarcRecord("bibsys/pub", marc.MustParseString(isbn9788203365133), testID())
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("ingestMarcRecord() mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("translated novel", func(t *testing.T) {
		const isbn8273504166 = `
			<record xmlns="http://www.loc.gov/MARC21/slim">
				<leader>01636cam a2200409 c 4500</leader>
				<controlfield tag="001">999410140454702201</controlfield>
				<controlfield tag="005">20211030183602.0</controlfield>
				<controlfield tag="007">ta</controlfield>
				<controlfield tag="007">cr||||||||||||</controlfield>
				<controlfield tag="008">150112s1994    no#||||| |||||000|1|nob|^</controlfield>
				<datafield ind1=" " ind2=" " tag="015">
					<subfield code="a">9405669</subfield>
					<subfield code="2">nbf</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="020">
					<subfield code="a">8203200168</subfield>
					<subfield code="q">ib.</subfield>
					<subfield code="c">Nkr 289.00</subfield>
					<subfield code="q">Aschehoug</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="020">
					<subfield code="a">8273504166</subfield>
					<subfield code="q">ib.</subfield>
					<subfield code="q">Dagens bok</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="035">
					<subfield code="a">941014045-47bibsys_network</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="035">
					<subfield code="a">(NO-TrBIB)941014045</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="035">
					<subfield code="a">(NO-TrBIB)092462138</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="035">
					<subfield code="a">(FHS-KS)14451</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="040">
					<subfield code="a">NO-OsNB</subfield>
					<subfield code="b">nob</subfield>
					<subfield code="e">katreg</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="041">
					<subfield code="h">swe</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="042">
					<subfield code="a">norbibl</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="044">
					<subfield code="c">no</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="080">
					<subfield code="a">839.7</subfield>
				</datafield>
				<datafield ind1="7" ind2="4" tag="082">
					<subfield code="a">839.73</subfield>
					<subfield code="q">NO-OsNB</subfield>
					<subfield code="2">4/nor</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="100">
					<subfield code="a">Ekman, Kerstin</subfield>
					<subfield code="d">1933-</subfield>
					<subfield code="0">(NO-TrBIB)90058909</subfield>
				</datafield>
				<datafield ind1="1" ind2="0" tag="245">
					<subfield code="a">Hendelser ved vann</subfield>
					<subfield code="c">Kerstin Ekman ; oversatt av Gunnel Malmström</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="246">
					<subfield code="a">Händelser vid vatten</subfield>
					<subfield code="i">Originaltittel</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="260">
					<subfield code="a">Oslo</subfield>
					<subfield code="b">Aschehoug</subfield>
					<subfield code="c">1994</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="300">
					<subfield code="a">446 s.</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="500">
					<subfield code="a">Fra 4. oppl. er Bokklubbens ISBN tillagt Bokklubben dagens bøker</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="500">
					<subfield code="a">Opplagshistorikk: 2.-3. oppl. 1994; 2. [i.e. 4.] oppl. 1997; 2. [i.e. 5. oppl.] 1998 (Nkr 310.00); [Nytt oppl.] 2000; [Nytt oppl.] 2001</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="533">
					<subfield code="a">Elektronisk reproduksjon</subfield>
					<subfield code="b">[Norge]</subfield>
					<subfield code="c">Nasjonalbiblioteket Digital</subfield>
					<subfield code="d">2007-12-21</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="653">
					<subfield code="a">skjønnlitteratur</subfield>
					<subfield code="a">roman</subfield>
					<subfield code="a">svensk-litteratur</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="700">
					<subfield code="a">Malmström, Gunnel</subfield>
					<subfield code="d">1921-2007</subfield>
					<subfield code="4">trl</subfield>
					<subfield code="0">(NO-TrBIB)90589668</subfield>
				</datafield>
				<datafield ind1="4" ind2="2" tag="856">
					<subfield code="3">Beskrivelse fra Forlagssentralen</subfield>
					<subfield code="u">http://content.bibsys.no/content/?type=descr_forlagssentr&amp;isbn=8203200168</subfield>
				</datafield>
				<datafield ind1="4" ind2="1" tag="856">
					<subfield code="3">Fulltekst</subfield>
					<subfield code="u">https://www.nb.no/search?q=oaiid:"oai:nb.bibsys.no:999410140454702202"&amp;mediatype=bøker</subfield>
					<subfield code="y">Nettbiblioteket</subfield>
					<subfield code="z">Søke-URL</subfield>
				</datafield>
				<datafield ind1="0" ind2="1" tag="852">
					<subfield code="a">47BIBSYS_UBO</subfield>
					<subfield code="6">999410140454702204</subfield>
					<subfield code="9">P</subfield>
				</datafield>
					<datafield ind1="0" ind2="1" tag="852">
					<subfield code="a">47BIBSYS_HIT</subfield>
					<subfield code="6">999410140454702210</subfield>
					<subfield code="9">P</subfield>
				</datafield>
				<datafield ind1="0" ind2="1" tag="852">
					<subfield code="a">47BIBSYS_UBIS</subfield>
					<subfield code="6">999410140454702208</subfield>
					<subfield code="9">P</subfield>
				</datafield>
				<datafield ind1="0" ind2="1" tag="852">
					<subfield code="a">47BIBSYS_UBTO</subfield>
					<subfield code="6">999410140454702205</subfield>
					<subfield code="9">P</subfield>
				</datafield>
				<datafield ind1="0" ind2="1" tag="852">
					<subfield code="a">47BIBSYS_UBIN</subfield>
					<subfield code="6">999410140454702211</subfield>
					<subfield code="9">P</subfield>
				</datafield>
				<datafield ind1="0" ind2="1" tag="852">
					<subfield code="a">47BIBSYS_HIB</subfield>
					<subfield code="6">999919807428902221</subfield>
					<subfield code="9">P</subfield>
				</datafield>
				<datafield ind1="0" ind2="1" tag="852">
					<subfield code="a">47BIBSYS_NB</subfield>
					<subfield code="6">999410140454702202</subfield>
					<subfield code="9">D</subfield>
					<subfield code="9">P</subfield>
				</datafield>
				<datafield ind1="0" ind2="1" tag="852">
					<subfield code="a">47BIBSYS_NTNU_UB</subfield>
					<subfield code="6">999410140454702203</subfield>
					<subfield code="9">P</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="901">
					<subfield code="a">90</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="913">
					<subfield code="a">Norbok</subfield>
					<subfield code="b">NB</subfield>
				</datafield>
			</record>
			`
		want := Ingestion{
			Resources: []sirkulator.Resource{
				{
					ID:    "t1",
					Type:  sirkulator.TypePublication,
					Label: "Kerstin Ekman - Hendelser ved vann (1994)",
					Data: sirkulator.Publication{
						Title:            "Hendelser ved vann",
						TitleOriginal:    "Händelser vid vatten",
						Language:         "nob",
						LanguageOriginal: "swe",
						//GenreForms: []string{"Romaner"}, // TODO
						Fiction:   true,
						Year:      1994,
						Publisher: "Aschehoug",
						NumPages:  446,
					},
				},
				{
					ID:    "t2",
					Type:  sirkulator.TypePerson,
					Label: "Kerstin Ekman (1933-)",
					Links: [][2]string{{"bibsys", "90058909"}},
					Data: sirkulator.Person{
						Name: "Kerstin Ekman",
						YearRange: sirkulator.YearRange{
							FromYear: 1933,
						},
					},
				},
				{
					ID:    "t3",
					Type:  sirkulator.TypePerson,
					Label: "Gunnel Malmström (1921-2007)",
					Links: [][2]string{{"bibsys", "90589668"}},
					Data: sirkulator.Person{
						Name: "Gunnel Malmström",
						YearRange: sirkulator.YearRange{
							FromYear: 1921,
							ToYear:   2007,
						},
					},
				},
			},
			Relations: []sirkulator.Relation{
				{
					FromID: "t1",
					ToID:   "t2",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "aut", "main_entry": true},
				},
				{
					FromID: "t1",
					ToID:   "t3",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "trl"},
				},
			},
			Reviews: []sirkulator.Relation{
				{
					FromID: "t1",
					Type:   "published_by",
					Data:   map[string]interface{}{"label": "Aschehoug"},
				},
			},
		}

		got, err := ingestMarcRecord("bibsys/pub", marc.MustParseString(isbn8273504166), testID())
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("ingestMarcRecord() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("reprint of textbook", func(t *testing.T) {
		// TODO detect duplicate author authority
		// one has kat3 level, one has kat1
		// but need to read bibsys/aut records to know that
		// - add "replaced_by" relation/link
		const isbn9788230021743 = `
			<record xmlns="http://www.loc.gov/MARC21/slim">
				<leader>02567cam a2200469 c 4500</leader>
				<controlfield tag="001">999921296219502201</controlfield>
				<controlfield tag="005">20211209092102.0</controlfield>
				<controlfield tag="007">ta</controlfield>
				<controlfield tag="008">210222s2021    no ab  e ||||||0| 0 nob|^</controlfield>
				<datafield ind1=" " ind2=" " tag="020">
					<subfield code="a">9788230021743</subfield>
					<subfield code="c">Nkr 490.00</subfield>
					<subfield code="q">innbundet</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="035">
					<subfield code="a">(NO-OsBA)0611226</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="035">
					<subfield code="a">oai:mlnb.bs.no:0611226</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="035">
					<subfield code="a">oai:bibbi.bs.no:0611226</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="040">
					<subfield code="a">NO-OsBA</subfield>
					<subfield code="b">nob</subfield>
					<subfield code="e">rda</subfield>
					<subfield code="d">NO-OsNB</subfield>
				</datafield>
				<datafield ind1="0" ind2="4" tag="082">
					<subfield code="a">581.6309481</subfield>
					<subfield code="2">23/nor</subfield>
					<subfield code="q">NO-OsNB</subfield>
				</datafield>
				<datafield ind1="0" ind2="4" tag="082">
					<subfield code="a">581.6309481</subfield>
					<subfield code="q">NO-OsBA</subfield>
					<subfield code="2">23/nor</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="100">
					<subfield code="a">Høeg, Ove Fredrik Arbo</subfield>
					<subfield code="d">1898-1993</subfield>
					<subfield code="0">(NO-TrBIB)1533887371289</subfield>
					<subfield code="4">aut</subfield>
				</datafield>
				<datafield ind1="1" ind2="0" tag="240">
					<subfield code="a">Planter og tradisjon</subfield>
				</datafield>
				<datafield ind1="1" ind2="0" tag="245">
					<subfield code="a">Planter og tradisjon</subfield>
					<subfield code="b">floraen i levende tale og tradisjon i Norge 1925-1973</subfield>
					<subfield code="c">Ove Arbo Høeg</subfield>
				</datafield>
				<datafield ind1=" " ind2="1" tag="264">
					<subfield code="a">[Oslo]</subfield>
					<subfield code="b">Norges sopp- og nyttevekstforbund</subfield>
					<subfield code="b">Nordic People and Plants</subfield>
					<subfield code="b">Universitetet i Oslo, Naturhistorisk museum</subfield>
					<subfield code="c">[2021]</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="300">
					<subfield code="a">VIII, 751 sider</subfield>
					<subfield code="b">illustrasjoner, kart</subfield>
					<subfield code="c">26 cm</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="336">
					<subfield code="a">tekst</subfield>
					<subfield code="0">http://rdaregistry.info/termList/RDAContentType/1020</subfield>
					<subfield code="2">rdaco</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="337">
					<subfield code="a">uformidlet</subfield>
					<subfield code="0">http://rdaregistry.info/termList/RDAMediaType/1007</subfield>
					<subfield code="2">rdamt</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="338">
					<subfield code="a">bind</subfield>
					<subfield code="0">http://rdaregistry.info/termList/RDACarrierType/1049</subfield>
					<subfield code="2">rdact</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="500">
					<subfield code="a">1. utgave Oslo : Universitetsforlaget, 1974</subfield>
				</datafield>
				<datafield ind1=" " ind2="7" tag="650">
					<subfield code="a">Nytteplanter</subfield>
					<subfield code="0">(NO-TrBIB)REAL002102</subfield>
					<subfield code="2">noubomn</subfield>
				</datafield>
				<datafield ind1=" " ind2="7" tag="650">
					<subfield code="a">Etnobotanikk</subfield>
					<subfield code="0">(NO-TrBIB)REAL009822</subfield>
					<subfield code="2">noubomn</subfield>
				</datafield>
				<datafield ind1=" " ind2="7" tag="651">
					<subfield code="a">Norge</subfield>
					<subfield code="0">(NO-TrBIB)REAL030753</subfield>
					<subfield code="2">noubomn</subfield>
				</datafield>
				<datafield ind1=" " ind2="7" tag="655">
					<subfield code="a">Populærvitenskap</subfield>
					<subfield code="0">(NO-TrBIB)REAL030121</subfield>
					<subfield code="2">noubomn</subfield>
				</datafield>
				<datafield ind1=" " ind2="7" tag="650">
					<subfield code="a">Etnobotanikk</subfield>
					<subfield code="z">Norge</subfield>
					<subfield code="0">(NO-OsBA)1200471</subfield>
					<subfield code="2">bibbi</subfield>
					<subfield code="9">nob</subfield>
				</datafield>
				<datafield ind1=" " ind2="7" tag="650">
					<subfield code="a">Etnobotanikk</subfield>
					<subfield code="z">Noreg</subfield>
					<subfield code="0">(NO-OsBA)1200471</subfield>
					<subfield code="2">bibbi</subfield>
					<subfield code="9">nno</subfield>
				</datafield>
				<datafield ind1="2" ind2="7" tag="650">
					<subfield code="a">Planter i folketroen</subfield>
					<subfield code="0">(NO-OsBA)1123309</subfield>
					<subfield code="2">bibbi</subfield>
					<subfield code="9">nob</subfield>
				</datafield>
				<datafield ind1="2" ind2="7" tag="650">
					<subfield code="a">Plantar i folketrua</subfield>
					<subfield code="0">(NO-OsBA)1123309</subfield>
					<subfield code="2">bibbi</subfield>
					<subfield code="9">nno</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="700">
					<subfield code="a">Høeg, Ove Arbo</subfield>
					<subfield code="d">1898-1993</subfield>
					<subfield code="t">Planter og tradisjon</subfield>
					<subfield code="0">(NO-TrBIB)90103766</subfield>
				</datafield>
				<datafield ind1="2" ind2=" " tag="710">
					<subfield code="a">Norges sopp- og nyttevekstforbund</subfield>
					<subfield code="0">(NO-TrBIB)5032677</subfield>
				</datafield>
				<datafield ind1="2" ind2=" " tag="710">
					<subfield code="a">Nordic People and Plants (prosjekt)</subfield>
					<subfield code="0">(NO-TrBIB)1614124828252</subfield>
				</datafield>
				<datafield ind1="2" ind2=" " tag="710">
					<subfield code="a">Universitetet i Oslo</subfield>
					<subfield code="b">Naturhistorisk museum</subfield>
					<subfield code="0">(NO-TrBIB)11071432</subfield>
				</datafield>
				<datafield ind1="4" ind2="2" tag="856">
					<subfield code="3">Originalt bilde</subfield>
					<subfield code="u">https://contents.bibs.aws.unit.no/files/images/original/3/4/9788230021743.jpg</subfield>
					<subfield code="q">image/jpeg</subfield>
				</datafield>
				<datafield ind1="4" ind2="2" tag="856">
					<subfield code="3">Forlagets beskrivelse (lang)</subfield>
					<subfield code="u">https://contents.bibs.aws.unit.no/content/?isbn=9788230021743</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="856">
					<subfield code="a">aja.bs.no</subfield>
					<subfield code="n">Biblioteksentralen, Oslo</subfield>
					<subfield code="q">image/jpeg</subfield>
					<subfield code="u">https://aja.bs.no/ad0ca9fa-1460-430b-a306-348a47f26437/cover/original.jpg</subfield>
					<subfield code="3">Omslagsbilde</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="856">
					<subfield code="a">aja.bs.no</subfield>
					<subfield code="n">Biblioteksentralen, Oslo</subfield>
					<subfield code="q">image/jpeg</subfield>
					<subfield code="u">https://aja.bs.no/ad0ca9fa-1460-430b-a306-348a47f26437/cover/thumbnail.jpg</subfield>
					<subfield code="3">Miniatyrbilde</subfield>
				</datafield>
				<datafield ind1="0" ind2="1" tag="852">
					<subfield code="a">47BIBSYS_UBO</subfield>
					<subfield code="6">999920326197902204</subfield>
					<subfield code="9">P</subfield>
				</datafield>
				<datafield ind1="0" ind2="1" tag="852">
					<subfield code="a">47BIBSYS_NB</subfield>
					<subfield code="6">999920120373402202</subfield>
					<subfield code="9">P</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="913">
					<subfield code="a">Norbok</subfield>
					<subfield code="b">NB</subfield>
				</datafield>
			</record>
			`
		want := Ingestion{
			Resources: []sirkulator.Resource{
				{
					ID:    "t1",
					Label: "Ove Fredrik Arbo Høeg - Planter og tradisjon: floraen i levende tale og tradisjon i Norge 1925-1973 (2021)",
					Type:  sirkulator.TypePublication,
					Data: sirkulator.Publication{
						Title:     "Planter og tradisjon",
						Subtitle:  "floraen i levende tale og tradisjon i Norge 1925-1973",
						Publisher: "Norges sopp- og nyttevekstforbund",
						Year:      2021,
						//YearFirst: 1974 // TODO from note field 500
						Language:   "nob",
						GenreForms: []string{"Populærvitenskap"},
						Nonfiction: true,
						NumPages:   751,
					},
				},
				{
					Type:  sirkulator.TypePerson,
					ID:    "t2",
					Label: "Ove Fredrik Arbo Høeg (1898-1993)",
					Links: [][2]string{{"bibsys", "1533887371289"}},
					Data: sirkulator.Person{
						Name: "Ove Fredrik Arbo Høeg",
						YearRange: sirkulator.YearRange{
							FromYear: 1898,
							ToYear:   1993,
						},
					},
				},
				{
					Type:  sirkulator.TypePerson,
					ID:    "t3",
					Label: "Ove Arbo Høeg (1898-1993)",
					Links: [][2]string{{"bibsys", "90103766"}},
					Data: sirkulator.Person{
						Name: "Ove Arbo Høeg",
						YearRange: sirkulator.YearRange{
							FromYear: 1898,
							ToYear:   1993,
						},
					},
				},
			},
			Relations: []sirkulator.Relation{
				{
					FromID: "t1",
					ToID:   "t2",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "aut", "main_entry": true},
				},
				{
					FromID: "t1",
					ToID:   "t3",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "aut"},
				},
			},
			Reviews: []sirkulator.Relation{
				{
					FromID: "t1",
					Type:   "published_by",
					Data:   map[string]interface{}{"label": "Norges sopp- og nyttevekstforbund"},
				},
			},
			Covers: []FileFetch{
				{
					ResourceID: "t1",
					URL:        "https://contents.bibs.aws.unit.no/files/images/original/3/4/9788230021743.jpg",
				},
				{
					ResourceID: "t1",
					URL:        "https://aja.bs.no/ad0ca9fa-1460-430b-a306-348a47f26437/cover/original.jpg",
				},
				{
					ResourceID: "t1",
					URL:        "https://aja.bs.no/ad0ca9fa-1460-430b-a306-348a47f26437/cover/thumbnail.jpg",
				},
			},
		}

		got, err := ingestMarcRecord("bibsys/pub", marc.MustParseString(isbn9788230021743), testID())
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("ingestMarcRecord() mismatch (-want +got):\n%s", diff)
		}
	})

}
