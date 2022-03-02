package etl

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/marc"
	"github.com/knakk/sirkulator/vocab"
)

func TestParseYearRange(t *testing.T) {
	tests := []struct {
		input string
		want  sirkulator.YearRange
	}{
		{"1981", sirkulator.YearRange{From: 1981}},
		{"  1981", sirkulator.YearRange{From: 1981}},
		{"1949-", sirkulator.YearRange{From: 1949}},
		{"1828-1906", sirkulator.YearRange{From: 1828, To: 1906}},
		{"1828 - 1906", sirkulator.YearRange{From: 1828, To: 1906}},
		{"1800-tallet", sirkulator.YearRange{From: 1800, To: 1900, Approx: true}},
		{"1500-tallet", sirkulator.YearRange{From: 1500, To: 1600, Approx: true}},
		{"fl. 1200-tallet", sirkulator.YearRange{From: 1200, To: 1300, Approx: true}},
		{"17. årh.", sirkulator.YearRange{From: 1600, To: 1700, Approx: true}},       // Technically 1601-1700
		{"16. årh.", sirkulator.YearRange{From: 1500, To: 1600, Approx: true}},       // Technically 1501-1600
		{"2. årh.", sirkulator.YearRange{From: 100, To: 200, Approx: true}},          // Technically 101-200
		{"2. årh. f.Kr.", sirkulator.YearRange{From: -200, To: -100, Approx: true}},  // Technically 200BC-101BC
		{"2. årh. f.Kr.?", sirkulator.YearRange{From: -200, To: -100, Approx: true}}, // Technically 200BC-101BC
		{"13th cent", sirkulator.YearRange{From: 1200, To: 1300, Approx: true}},      // Technically 1201-1300
		{"16th cent", sirkulator.YearRange{From: 1500, To: 1600, Approx: true}},      // Technically 1501-1600
		{"382-336 f.Kr.", sirkulator.YearRange{From: -382, To: -336}},
		{"død 1836", sirkulator.YearRange{To: 1836}},
		{"d. 1650", sirkulator.YearRange{To: 1650}},
		{"d. ca. 1480", sirkulator.YearRange{To: 1480, Approx: true}},
		{"d. 514 f.Kr.", sirkulator.YearRange{To: -514}},
		{"-1755", sirkulator.YearRange{To: 1755}},
		{"--1989", sirkulator.YearRange{To: 1989}},
		{"b. 1883", sirkulator.YearRange{From: 1883}},
		{"f. 1891", sirkulator.YearRange{From: 1891}},
		{"f. ca 1685", sirkulator.YearRange{From: 1685, Approx: true}},
		{"(1961- )", sirkulator.YearRange{From: 1961}},
		{"[1774-1857]", sirkulator.YearRange{From: 1774, To: 1857}},
		{"ca. 1030-ca. 1112", sirkulator.YearRange{From: 1030, To: 1112, Approx: true}},
	}
	// TODO cases:
	//  2./3. årh.
	//  4./3. årh. f.Kr.
	//  virksom 1849
	//  virksom 18. årh.
	//  virksom omkr. 1840
	//  Virksom ca. 1761
	//  virksom 1685-1711
	//  aktiv på 1000-tallet
	//  19-? | 17-?
	//  19??
	//  1980?
	//  1700-1800-tallet
	//  1700-tallet-1800-tallet
	//  1700-tallet?
	//  19
	//  f. 20. årh. | f. 18. årh.
	//  19. årh.-20. årh.?
	//  1871-?
	//  1907?-1979 | 1181?-1246
	//  1960-....
	//  1945-03-25
	//  ca 1705
	//  1872-ca. 1950?
	//  1862-ca. 1930
	//  1881 [eller 1889]-1943
	//  43 B.C.-17 or 18 A.D

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

func TestParseYear(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		// good:
		{"1981", 1981},
		{"1981 ", 1981},
		{" 1981", 1981},
		{"1981.", 1981},
		{"cop. 1981", 1981},
		{"c2004", 2004},
		{"c2012.", 2012},
		{"[1996]", 1996},
		{"©2004", 2004},
		{"[1986?]", 1986}, // ?
		// no-good:
		{"[19-?]", 0},
		{"[198-?]", 0},
		{"[u.å.]", 0},
		{"[s.a.]", 0},
		{"c1993-2010 [2010]", 0},
		{"20060101", 0},
		{"2012, c2013.", 0},
	}

	for _, test := range tests {
		if got := parseYear(test.input); got != test.want {
			t.Errorf("parseYear(%q): got %v; want %v", test.input, got, test.want)
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
					Links: [][2]string{{"isbn", "9788203365133"}},
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
					Label: "Liv Køltzow (1945–)",
					Links: [][2]string{{"bibsys", "90086277"}},
					Data: sirkulator.Person{
						Name: "Liv Køltzow",
						YearRange: sirkulator.YearRange{
							From: 1945,
						},
					},
				},
				{
					ID:    "t3",
					Type:  sirkulator.TypePerson,
					Label: "Hans Petter Blad (1962–)",
					Links: [][2]string{{"bibsys", "90916002"}},
					Data: sirkulator.Person{
						Name: "Hans Petter Blad",
						YearRange: sirkulator.YearRange{
							From: 1962,
						},
					},
				},
				{
					ID:    "t4",
					Type:  sirkulator.TypePerson,
					Label: "Kaja Schjerven Mollerin (1980–)",
					Links: [][2]string{{"bibsys", "6088516"}},
					Data: sirkulator.Person{
						Name: "Kaja Schjerven Mollerin",
						YearRange: sirkulator.YearRange{
							From: 1980,
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
					Links: [][2]string{{"isbn", "8203200168"}, {"isbn", "8273504166"}},
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
					Label: "Kerstin Ekman (1933–)",
					Links: [][2]string{{"bibsys", "90058909"}},
					Data: sirkulator.Person{
						Name: "Kerstin Ekman",
						YearRange: sirkulator.YearRange{
							From: 1933,
						},
					},
				},
				{
					ID:    "t3",
					Type:  sirkulator.TypePerson,
					Label: "Gunnel Malmström (1921–2007)",
					Links: [][2]string{{"bibsys", "90589668"}},
					Data: sirkulator.Person{
						Name: "Gunnel Malmström",
						YearRange: sirkulator.YearRange{
							From: 1921,
							To:   2007,
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
					<subfield code="u">https://aja.bs.no/ad0ca9fa-1460-430b-a306-348a47f26437/cover/thumbnail.jpg</subfield>
					<subfield code="3">Miniatyrbilde</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="856">
					<subfield code="a">aja.bs.no</subfield>
					<subfield code="n">Biblioteksentralen, Oslo</subfield>
					<subfield code="q">image/jpeg</subfield>
					<subfield code="u">https://aja.bs.no/ad0ca9fa-1460-430b-a306-348a47f26437/cover/original.jpg</subfield>
					<subfield code="3">Omslagsbilde</subfield>
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
					Links: [][2]string{{"isbn", "9788230021743"}},
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
					Label: "Ove Fredrik Arbo Høeg (1898–1993)",
					Links: [][2]string{{"bibsys", "1533887371289"}},
					Data: sirkulator.Person{
						Name: "Ove Fredrik Arbo Høeg",
						YearRange: sirkulator.YearRange{
							From: 1898,
							To:   1993,
						},
					},
				},
				/*{
					Type:  sirkulator.TypePerson,
					ID:    "t3",
					Label: "Ove Arbo Høeg (1898–1993)",
					Links: [][2]string{{"bibsys", "90103766"}},
					Data: sirkulator.Person{
						Name: "Ove Arbo Høeg",
						YearRange: sirkulator.YearRange{
							From: 1898,
							To:   1993,
						},
					},
				},
				{
					Type:  sirkulator.TypeCorporation,
					ID:    "t4",
					Label: "Norges sopp- og nyttevekstforbund",
					Links: [][2]string{{"bibsys", "5032677"}},
				},
				{
					Type:  sirkulator.TypeCorporation,
					ID:    "t5",
					Label: "Nordic People and Plants (prosjekt)",
					Links: [][2]string{{"bibsys", "1614124828252"}},
				},
				{
					Type:  sirkulator.TypeCorporation,
					ID:    "t6",
					Label: "Universitetet i Oslo",
					Links: [][2]string{{"bibsys", "11071432"}},
				},*/
			},
			Relations: []sirkulator.Relation{
				{
					FromID: "t1",
					ToID:   "t2",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "aut", "main_entry": true},
				},
				/*{
					FromID: "t1",
					ToID:   "t3",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "aut"},
				},
				{
					FromID: "t1",
					ToID:   "t4",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": string("aut")},
				},
				{
					FromID: "t1",
					ToID:   "t5",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": string("aut")},
				},
				{
					FromID: "t1",
					ToID:   "t6",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": string("aut")},
				},*/
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

	t.Run("non-fiction with many contributors", func(t *testing.T) {
		const isbn9788253043203 = `
			<record xmlns="http://www.loc.gov/MARC21/slim">
				<leader>0194922   220038500 4500</leader>
				<controlfield tag="001">999921641921002201</controlfield>
				<controlfield tag="005">20220131131755.0</controlfield>
				<controlfield tag="007">ta</controlfield>
				<controlfield tag="008">220105s2022    no a   e ||||||0| 0 nob|</controlfield>
				<datafield ind1=" " ind2=" " tag="020">
					<subfield code="a">9788253043203</subfield>
					<subfield code="c">Nkr 499.00</subfield>
					<subfield code="q">innbundet</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="035">
					<subfield code="a">(NO-OsBA)0646223</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="035">
					<subfield code="a">oai:bibbi.bs.no:0646223</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="040">
					<subfield code="a">NO-OsBA</subfield>
					<subfield code="b">nob</subfield>
					<subfield code="e">rda</subfield>
					<subfield code="d">NO-OsNB</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="080">
					<subfield code="a">727.8(481.13)</subfield>
				</datafield>
				<datafield ind1="0" ind2="4" tag="082">
					<subfield code="a">727.8244821</subfield>
					<subfield code="q">NO-OsBA</subfield>
					<subfield code="2">23/nor</subfield>
				</datafield>
				<datafield ind1="0" ind2="0" tag="245">
					<subfield code="a">Deichman Bjørvika</subfield>
					<subfield code="b">Lundhagem og Atelier Oslo arkitekter</subfield>
					<subfield code="c">redaktører Lars Müller og arkitektene ; med bidrag fra Niklas Maak, Elif Shafak, Liv Sæteren ; fotoessays av Einar Aslaksen, Iwan Baan, Hélène Binet ; oversettere: Jan Christopher Næss og Lene Stokseth</subfield>
				</datafield>
				<datafield ind1=" " ind2="1" tag="264">
					<subfield code="a">Oslo</subfield>
					<subfield code="b">Pax forlag</subfield>
					<subfield code="c">[2022]</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="300">
					<subfield code="a">271 sider</subfield>
					<subfield code="b">illustrasjoner i farger</subfield>
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
				<datafield ind1=" " ind2="7" tag="650">
					<subfield code="a">Bibliotekbygninger</subfield>
					<subfield code="z">Oslo</subfield>
					<subfield code="2">tekord</subfield>
				</datafield>
				<datafield ind1="2" ind2="7" tag="610">
					<subfield code="a">Deichman Bjørvika</subfield>
					<subfield code="0">(NO-TrBIB)1642068353945</subfield>
					<subfield code="2">bare</subfield>
				</datafield>
				<datafield ind1="2" ind2="7" tag="650">
					<subfield code="a">Folkebibliotek</subfield>
					<subfield code="g">arkitektur</subfield>
					<subfield code="z">Oslo</subfield>
					<subfield code="0">(NO-OsBA)1263818</subfield>
					<subfield code="2">bibbi</subfield>
					<subfield code="9">nob</subfield>
				</datafield>
				<datafield ind1="2" ind2="7" tag="650">
					<subfield code="a">Folkebibliotek</subfield>
					<subfield code="g">arkitektur</subfield>
					<subfield code="z">Oslo</subfield>
					<subfield code="0">(NO-OsBA)1263818</subfield>
					<subfield code="2">bibbi</subfield>
					<subfield code="9">nno</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="700">
					<subfield code="a">Müller, Lars</subfield>
					<subfield code="4">edt</subfield>
					<subfield code="0">(NO-TrBIB)90961231</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="700">
					<subfield code="a">Shafak, Elif</subfield>
					<subfield code="d">1971-</subfield>
					<subfield code="4">aut</subfield>
					<subfield code="0">(NO-TrBIB)8035652</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="700">
					<subfield code="a">Maak, Niklas</subfield>
					<subfield code="d">1972-</subfield>
					<subfield code="4">aut</subfield>
					<subfield code="0">(NO-TrBIB)12010078</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="700">
					<subfield code="a">Sæteren, Liv</subfield>
					<subfield code="4">aut</subfield>
					<subfield code="0">(NO-TrBIB)9009005</subfield>
				</datafield>
				<datafield ind1="2" ind2=" " tag="710">
					<subfield code="a">Lund Hagem arkitekter AS</subfield>
				</datafield>
				<datafield ind1="2" ind2=" " tag="710">
					<subfield code="a">Atelier Oslo</subfield>
					<subfield code="0">(NO-TrBIB)12073195</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="856">
					<subfield code="a">aja.bs.no</subfield>
					<subfield code="q">image/jpeg</subfield>
					<subfield code="u">https://media.aja.bs.no/cd6935ab-1a2b-4e29-ac07-a3ba5d753d89/cover/thumbnail.jpg</subfield>
					<subfield code="3">Miniatyrbilde</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="856">
					<subfield code="a">aja.bs.no</subfield>
					<subfield code="q">image/jpeg</subfield>
					<subfield code="u">https://media.aja.bs.no/cd6935ab-1a2b-4e29-ac07-a3ba5d753d89/cover/original.jpg</subfield>
					<subfield code="3">Omslagsbilde</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="913">
					<subfield code="a">Norbok</subfield>
					<subfield code="b">NB</subfield>
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
					Label: "Deichman Bjørvika: Lundhagem og Atelier Oslo arkitekter (2022)",
					Type:  sirkulator.TypePublication,
					Links: [][2]string{{"isbn", "9788253043203"}},
					Data: sirkulator.Publication{
						Title:     "Deichman Bjørvika",
						Subtitle:  "Lundhagem og Atelier Oslo arkitekter",
						Publisher: "Pax forlag",
						Year:      2022,
						//YearFirst: 2022
						Language:   "nob",
						Nonfiction: true,
						NumPages:   271,
					},
				},
				{
					Type:  sirkulator.TypeCorporation,
					ID:    "t2",
					Label: "Deichman Bjørvika",
					Links: [][2]string{{"bibsys", "1642068353945"}},
				},
				{
					Type:  sirkulator.TypePerson,
					ID:    "t3",
					Label: "Lars Müller",
					Links: [][2]string{{"bibsys", "90961231"}},
					Data: sirkulator.Person{
						Name: "Lars Müller",
					},
				},
				{
					Type:  sirkulator.TypePerson,
					ID:    "t4",
					Label: "Elif Shafak (1971–)",
					Links: [][2]string{{"bibsys", "8035652"}},
					Data: sirkulator.Person{
						Name: "Elif Shafak",
						YearRange: sirkulator.YearRange{
							From: 1971,
						},
					},
				},
				{
					Type:  sirkulator.TypePerson,
					ID:    "t5",
					Label: "Niklas Maak (1972–)",
					Links: [][2]string{{"bibsys", "12010078"}},
					Data: sirkulator.Person{
						Name: "Niklas Maak",
						YearRange: sirkulator.YearRange{
							From: 1972,
						},
					},
				},
				{
					Type:  sirkulator.TypePerson,
					ID:    "t6",
					Label: "Liv Sæteren",
					Links: [][2]string{{"bibsys", "9009005"}},
					Data: sirkulator.Person{
						Name: "Liv Sæteren",
					},
				},
				/*{
					Type:  sirkulator.TypeCorporation,
					ID:    "t7",
					Label: "Lund Hagem arkitekter AS",
				},
				{
					Type:  sirkulator.TypeCorporation,
					ID:    "t8",
					Label: "Atelier Oslo",
					Links: [][2]string{{"bibsys", "12073195"}},
				},*/
			},
			Relations: []sirkulator.Relation{
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
					Data:   map[string]interface{}{"role": "aut"},
				},
				{
					FromID: "t1",
					ToID:   "t5",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "aut"},
				},
				{
					FromID: "t1",
					ToID:   "t6",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "aut"},
				},
				/*{
					FromID: "t1",
					ToID:   "t7",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": string("aut")},
				},
				{
					FromID: "t1",
					ToID:   "t8",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": string("aut")},
				},*/
			},
			Reviews: []sirkulator.Relation{
				{
					FromID: "t1",
					Type:   "published_by",
					Data:   map[string]interface{}{"label": "Pax forlag"},
				},
			},
			Covers: []FileFetch{
				{
					ResourceID: "t1",
					URL:        "https://media.aja.bs.no/cd6935ab-1a2b-4e29-ac07-a3ba5d753d89/cover/original.jpg",
				},
				{
					ResourceID: "t1",
					URL:        "https://media.aja.bs.no/cd6935ab-1a2b-4e29-ac07-a3ba5d753d89/cover/thumbnail.jpg",
				},
			},
		}

		got, err := ingestMarcRecord("bibsys/pub", marc.MustParseString(isbn9788253043203), testID())
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("ingestMarcRecord() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("book in a publisher series", func(t *testing.T) {
		const isbn9788205560130 = `
			<record xmlns="http://www.loc.gov/MARC21/slim">
				<leader>02682nam  2200481 c 4500</leader>
				<controlfield tag="001">999921603445902201</controlfield>
				<controlfield tag="005">20220223113839.0</controlfield>
				<controlfield tag="007">ta</controlfield>
				<controlfield tag="008">211029s2022    no     j ||||||0| 1 nob|^</controlfield>
				<datafield ind1=" " ind2=" " tag="020">
					<subfield code="a">9788205560130</subfield>
					<subfield code="c">Nkr 279.00</subfield>
					<subfield code="q">innbundet</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="035">
					<subfield code="a">(NO-OsBA)0638924</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="035">
					<subfield code="a">oai:bibbi.bs.no:0638924</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="040">
					<subfield code="a">NO-OsBA</subfield>
					<subfield code="b">nob</subfield>
					<subfield code="e">rda</subfield>
					<subfield code="d">NO-OsNB</subfield>
				</datafield>
				<datafield ind1="0" ind2="4" tag="082">
					<subfield code="a">839.8238</subfield>
					<subfield code="q">NO-OsBA</subfield>
					<subfield code="2">23/nor</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="100">
					<subfield code="a">Ingvaldsen, Bjørn</subfield>
					<subfield code="d">1962-</subfield>
					<subfield code="0">(NO-TrBIB)90829580</subfield>
					<subfield code="4">aut</subfield>
				</datafield>
				<datafield ind1="1" ind2="0" tag="240">
					<subfield code="a">Når noen klikker i vinkel</subfield>
				</datafield>
				<datafield ind1="1" ind2="0" tag="245">
					<subfield code="a">Når noen klikker i vinkel</subfield>
					<subfield code="c">Bjørn Ingvaldsen</subfield>
				</datafield>
				<datafield ind1=" " ind2="1" tag="264">
					<subfield code="a">Oslo</subfield>
					<subfield code="b">Gyldendal</subfield>
					<subfield code="c">2022</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="300">
					<subfield code="a">122 sider</subfield>
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
				<datafield ind1=" " ind2=" " tag="385">
					<subfield code="a">9-10 år</subfield>
					<subfield code="m">Aldersgruppe</subfield>
					<subfield code="0">https://schema.nb.no/Bibliographic/Values/TG1003</subfield>
					<subfield code="2">nortarget</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="385">
					<subfield code="a">11-12 år</subfield>
					<subfield code="m">Aldersgruppe</subfield>
					<subfield code="0">https://schema.nb.no/Bibliographic/Values/TG1004</subfield>
					<subfield code="2">nortarget</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="385">
					<subfield code="a">Leselig skrift</subfield>
					<subfield code="m">Gruppe med spesielle behov</subfield>
					<subfield code="0">https://schema.nb.no/Bibliographic/Values/TG1009</subfield>
					<subfield code="2">nortarget</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="490">
					<subfield code="a">Søstrene Proxy blogger om verden</subfield>
					<subfield code="v">1</subfield>
				</datafield>
				<datafield ind1=" " ind2=" " tag="520">
					<subfield code="a">Søstrene Mercedes på 13 og Lada på nesten 12 tester ut livet som bloggere og influensere. I den alderen vil man jo gjerne fremstå som eldre enn man er, og mamma er litt bekymret når døtrene setter i gang med kamera og gode planer. Og jentene lærer at det fort kan bli litt pinlig og ganske morsomt. Humor for mellomtrinnet. Omtalen er utarbeidet av BS.</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="521">
					<subfield code="a">Alder: 9-13 år</subfield>
				</datafield>
				<datafield ind1="3" ind2=" " tag="521">
					<subfield code="a">God, leselig skrift</subfield>
				</datafield>
				<datafield ind1=" " ind2="7" tag="655">
					<subfield code="a">Romaner</subfield>
					<subfield code="0">https://id.nb.no/vocabulary/ntsf/258</subfield>
					<subfield code="2">ntsf</subfield>
					<subfield code="9">nob</subfield>
				</datafield>
				<datafield ind1=" " ind2="7" tag="655">
					<subfield code="a">Romanar</subfield>
					<subfield code="0">https://id.nb.no/vocabulary/ntsf/258</subfield>
					<subfield code="2">ntsf</subfield>
					<subfield code="9">nno</subfield>
				</datafield>
				<datafield ind1=" " ind2="7" tag="655">
					<subfield code="a">Humor</subfield>
					<subfield code="0">https://id.nb.no/vocabulary/ntsf/127</subfield>
					<subfield code="2">ntsf</subfield>
					<subfield code="9">nob</subfield>
				</datafield>
				<datafield ind1=" " ind2="7" tag="655">
					<subfield code="a">Humor</subfield>
					<subfield code="0">https://id.nb.no/vocabulary/ntsf/127</subfield>
					<subfield code="2">ntsf</subfield>
					<subfield code="9">nno</subfield>
				</datafield>
				<datafield ind1="2" ind2="7" tag="650">
					<subfield code="a">Blogging</subfield>
					<subfield code="0">(NO-OsBA)1140916</subfield>
					<subfield code="2">bibbi</subfield>
					<subfield code="9">nob</subfield>
				</datafield>
				<datafield ind1="2" ind2="7" tag="650">
					<subfield code="a">Blogging</subfield>
					<subfield code="0">(NO-OsBA)1140916</subfield>
					<subfield code="2">bibbi</subfield>
					<subfield code="9">nno</subfield>
				</datafield>
				<datafield ind1="2" ind2="7" tag="650">
					<subfield code="a">Søstre</subfield>
					<subfield code="0">(NO-OsBA)1128323</subfield>
					<subfield code="2">bibbi</subfield>
					<subfield code="9">nob</subfield>
				</datafield>
				<datafield ind1="2" ind2="7" tag="650">
					<subfield code="a">Systrer</subfield>
					<subfield code="0">(NO-OsBA)1128323</subfield>
					<subfield code="2">bibbi</subfield>
					<subfield code="9">nno</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="700">
					<subfield code="a">Bergesen, Anders</subfield>
					<subfield code="d">1976-</subfield>
					<subfield code="4">bjd</subfield>
					<subfield code="0">(NO-TrBIB)10007330</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="700">
					<subfield code="a">Ingvaldsen, Bjørn</subfield>
					<subfield code="d">1962-</subfield>
					<subfield code="t">Når noen klikker i vinkel</subfield>
					<subfield code="0">(NO-TrBIB)90829580</subfield>
				</datafield>
				<datafield ind1="1" ind2=" " tag="800">
					<subfield code="a">Ingvaldsen, Bjørn</subfield>
					<subfield code="d">1962-</subfield>
					<subfield code="t">Søstrene Proxy blogger om verden</subfield>
					<subfield code="l">Norsk</subfield>
					<subfield code="v">1</subfield>
					<subfield code="0">(NO-TrBIB)90829580</subfield>
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
					Label: "Bjørn Ingvaldsen - Når noen klikker i vinkel (2022)",
					Type:  sirkulator.TypePublication,
					Links: [][2]string{{"isbn", "9788205560130"}},
					Data: sirkulator.Publication{
						Title:      "Når noen klikker i vinkel",
						Publisher:  "Gyldendal",
						Series:     []string{"Søstrene Proxy blogger om verden"},
						Year:       2022,
						Language:   "nob",
						GenreForms: []string{"Romaner", "Humor"},
						Audiences:  []vocab.Audience{vocab.TG1003, vocab.TG1004, vocab.TG1009},
						Fiction:    true,
						NumPages:   122,
					},
				},
				{
					Type:  sirkulator.TypePerson,
					ID:    "t2",
					Label: "Bjørn Ingvaldsen (1962–)",
					Data: sirkulator.Person{
						YearRange: sirkulator.YearRange{From: 1962},
						Name:      "Bjørn Ingvaldsen"},
					Links: [][2]string{{"bibsys", "90829580"}},
				},
				{
					Type:  sirkulator.TypePerson,
					ID:    "t3",
					Label: "Anders Bergesen (1976–)",
					Links: [][2]string{{"bibsys", "10007330"}},
					Data: sirkulator.Person{
						Name:      "Anders Bergesen",
						YearRange: sirkulator.YearRange{From: 1976},
					},
				},
			},
			Relations: []sirkulator.Relation{
				{
					FromID: "t1",
					ToID:   "t2",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"main_entry": true, "role": "aut"},
				},
				{
					FromID: "t1",
					ToID:   "t3",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "bjd"},
				},
			},
			Reviews: []sirkulator.Relation{
				{
					FromID: "t1",
					Type:   "published_by",
					Data:   map[string]interface{}{"label": "Gyldendal"},
				},
				{
					FromID: "t1",
					Type:   "in_series",
					Data: map[string]interface{}{
						"label":     "Søstrene Proxy blogger om verden",
						"number":    1,
						"publisher": "Gyldendal",
						//	"author": "Bjørn Ingvaldsen",
					},
				},
			},
		}

		got, err := ingestMarcRecord("bibsys/pub", marc.MustParseString(isbn9788205560130), testID())
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("ingestMarcRecord() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("book with audience codes", func(t *testing.T) {
		const isbn9788202527921 = `<record xmlns="http://www.loc.gov/MARC21/slim">
		<leader>03117nam a2200529 c 4500</leader>
		<controlfield tag="001">999920253627802201</controlfield>
		<controlfield tag="005">20211104172936.0</controlfield>
		<controlfield tag="007">ta</controlfield>
		<controlfield tag="008">170118s2017    no a   j |||||00| j nob|^</controlfield>
		<datafield ind1=" " ind2=" " tag="020">
		  <subfield code="a">978-82-02-52792-1</subfield>
		  <subfield code="q">ib.</subfield>
		  <subfield code="c">Nkr 169.00</subfield>
		</datafield>
		<datafield ind1=" " ind2=" " tag="035">
		  <subfield code="a">(NO-OsBAS)150262149</subfield>
		</datafield>
		<datafield ind1=" " ind2=" " tag="040">
		  <subfield code="a">NO-OsBAS</subfield>
		  <subfield code="b">nob</subfield>
		  <subfield code="d">NO-OsNB</subfield>
		  <subfield code="e">katreg</subfield>
		</datafield>
		<datafield ind1="0" ind2="4" tag="082">
		  <subfield code="a">839.823</subfield>
		  <subfield code="q">NO-OsBAS</subfield>
		  <subfield code="2">23/nor</subfield>
		</datafield>
		<datafield ind1="1" ind2=" " tag="100">
		  <subfield code="a">Bringsværd, Tor Åge</subfield>
		  <subfield code="d">1939-</subfield>
		  <subfield code="0">(NO-TrBIB)90058926</subfield>
		  <subfield code="4">aut</subfield>
		</datafield>
		<datafield ind1="1" ind2="4" tag="245">
		  <subfield code="a">Den glemte byen</subfield>
		  <subfield code="c">Tor Åge Bringsværd ; illustrert av Haakon Lie</subfield>
		</datafield>
		<datafield ind1=" " ind2=" " tag="260">
		  <subfield code="a">[Oslo]</subfield>
		  <subfield code="b">Cappelen Damm</subfield>
		  <subfield code="c">2017</subfield>
		</datafield>
		<datafield ind1=" " ind2=" " tag="300">
		  <subfield code="a">53 s.</subfield>
		  <subfield code="b">kol. ill.</subfield>
		  <subfield code="c">24 cm</subfield>
		</datafield>
		<datafield ind1=" " ind2=" " tag="336">
		  <subfield code="a">text</subfield>
		  <subfield code="b">txt</subfield>
		  <subfield code="2">rdacontent</subfield>
		</datafield>
		<datafield ind1=" " ind2=" " tag="337">
		  <subfield code="a">unmediated</subfield>
		  <subfield code="b">n</subfield>
		  <subfield code="2">rdamedia</subfield>
		</datafield>
		<datafield ind1=" " ind2=" " tag="338">
		  <subfield code="a">volume</subfield>
		  <subfield code="b">nc</subfield>
		  <subfield code="2">rdacarrier</subfield>
		</datafield>
		<datafield ind1=" " ind2=" " tag="385">
		  <subfield code="b">b</subfield>
		  <subfield code="m">Age group</subfield>
		  <subfield code="2">nortarget</subfield>
		</datafield>
		<datafield ind1=" " ind2=" " tag="385">
		  <subfield code="b">ta</subfield>
		  <subfield code="m">Specialized group</subfield>
		  <subfield code="2">nortarget</subfield>
		</datafield>
		<datafield ind1="1" ind2=" " tag="490">
		  <subfield code="a">Ulvegutten Tal</subfield>
		  <subfield code="v">8</subfield>
		</datafield>
		<datafield ind1="1" ind2=" " tag="490">
		  <subfield code="a">Min første leseløve</subfield>
		</datafield>
		<datafield ind1=" " ind2=" " tag="520">
		  <subfield code="a">Den åttende boka i serien om Tal, Nea og Shita er minst like spennende som de foregående! Nea er murr-stammens nye Tigerdronningen - selv om hun bare er ni år! I det siste har de hørt grusomme lyder fra Den glemte byen, og Nea bestemmer at de må finne ut hva som skjer. Tal og Nea blir med de modige jegerne inn i jungelen. Når de nærmer seg Den glemte byen, blir hylene høyere og mer skremmende - og plutselig ser de en stor, skummel skygge som nærmer seg...Tor Åge Bringsværd og Haakon Lie går ikke tomme for spennende og farlige eventyr for ulvegutten Tal og mammut-jenta Nea.</subfield>
		</datafield>
		<datafield ind1=" " ind2="7" tag="648">
		  <subfield code="a">Steinalderen</subfield>
		  <subfield code="2">bokbas</subfield>
		  <subfield code="9">nob</subfield>
		</datafield>
		<datafield ind1=" " ind2="7" tag="648">
		  <subfield code="a">Steinalderen</subfield>
		  <subfield code="2">bokbas</subfield>
		  <subfield code="9">nno</subfield>
		</datafield>
		<datafield ind1=" " ind2="7" tag="655">
		  <subfield code="a">Romaner</subfield>
		  <subfield code="2">bokbas</subfield>
		  <subfield code="9">nob</subfield>
		</datafield>
		<datafield ind1=" " ind2="7" tag="655">
		  <subfield code="a">Romanar</subfield>
		  <subfield code="2">bokbas</subfield>
		  <subfield code="9">nno</subfield>
		</datafield>
		<datafield ind1=" " ind2="7" tag="655">
		  <subfield code="a">Lettlest</subfield>
		  <subfield code="2">bokbas</subfield>
		  <subfield code="9">nob</subfield>
		</datafield>
		<datafield ind1=" " ind2="7" tag="655">
		  <subfield code="a">Lettlest</subfield>
		  <subfield code="2">bokbas</subfield>
		  <subfield code="9">nno</subfield>
		</datafield>
		<datafield ind1=" " ind2="7" tag="655">
		  <subfield code="a">Spenning</subfield>
		  <subfield code="2">bokbas</subfield>
		  <subfield code="9">nob</subfield>
		</datafield>
		<datafield ind1=" " ind2="7" tag="655">
		  <subfield code="a">Spenning</subfield>
		  <subfield code="2">bokbas</subfield>
		  <subfield code="9">nno</subfield>
		</datafield>
		<datafield ind1=" " ind2="7" tag="655">
		  <subfield code="a">Historisk litteratur</subfield>
		  <subfield code="2">bokbas</subfield>
		  <subfield code="9">nob</subfield>
		</datafield>
		<datafield ind1=" " ind2="7" tag="655">
		  <subfield code="a">Historisk litteratur</subfield>
		  <subfield code="2">bokbas</subfield>
		  <subfield code="9">nno</subfield>
		</datafield>
		<datafield ind1="1" ind2=" " tag="700">
		  <subfield code="a">Lie, Haakon</subfield>
		  <subfield code="d">1991-</subfield>
		  <subfield code="0">(NO-TrBIB)13037663</subfield>
		  <subfield code="4">ill</subfield>
		</datafield>
		<datafield ind1="1" ind2=" " tag="800">
		  <subfield code="a">Bringsværd, Tor Åge</subfield>
		  <subfield code="t">Ulvegutten Tal</subfield>
		  <subfield code="v">8</subfield>
		</datafield>
		<datafield ind1=" " ind2="0" tag="830">
		  <subfield code="a">Min første leseløve</subfield>
		</datafield>
	  </record>`

		want := Ingestion{
			Resources: []sirkulator.Resource{
				{
					ID:    "t1",
					Label: "Tor Åge Bringsværd - Den glemte byen (2017)",
					Type:  sirkulator.TypePublication,
					Links: [][2]string{{"isbn", "9788202527921"}},
					Data: sirkulator.Publication{
						Title:      "Den glemte byen",
						Publisher:  "Cappelen Damm",
						Series:     []string{"Ulvegutten Tal", "Min første leseløve"},
						Year:       2017,
						Language:   "nob",
						GenreForms: []string{"Romaner", "Lettlest", "Spenning", "Historisk litteratur"},
						Audiences:  []vocab.Audience{vocab.TG1002, vocab.TG1006},
						Fiction:    true,
						NumPages:   53,
					},
				},
				{
					Type:  sirkulator.TypePerson,
					ID:    "t2",
					Label: "Tor Åge Bringsværd (1939–)",
					Data: sirkulator.Person{
						YearRange: sirkulator.YearRange{From: 1939},
						Name:      "Tor Åge Bringsværd"},
					Links: [][2]string{{"bibsys", "90058926"}},
				},
				{
					Type:  sirkulator.TypePerson,
					ID:    "t3",
					Label: "Haakon Lie (1991–)",
					Links: [][2]string{{"bibsys", "13037663"}},
					Data: sirkulator.Person{
						Name:      "Haakon Lie",
						YearRange: sirkulator.YearRange{From: 1991},
					},
				},
			},
			Relations: []sirkulator.Relation{
				{
					FromID: "t1",
					ToID:   "t2",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"main_entry": true, "role": "aut"},
				},
				{
					FromID: "t1",
					ToID:   "t3",
					Type:   "has_contributor",
					Data:   map[string]interface{}{"role": "ill"},
				},
			},
			Reviews: []sirkulator.Relation{
				{
					FromID: "t1",
					Type:   "published_by",
					Data:   map[string]interface{}{"label": "Cappelen Damm"},
				},
				{
					FromID: "t1",
					Type:   "in_series",
					Data: map[string]interface{}{
						"label":     "Ulvegutten Tal",
						"number":    8,
						"publisher": "Cappelen Damm",
					},
				},
				{
					FromID: "t1",
					Type:   "in_series",
					Data: map[string]interface{}{
						"label":     "Min første leseløve",
						"publisher": "Cappelen Damm",
					},
				},
			},
		}

		got, err := ingestMarcRecord("bibsys/pub", marc.MustParseString(isbn9788202527921), testID())
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("ingestMarcRecord() mismatch (-want +got):\n%s", diff)
		}
	})

}

func TestPersonFromAuthority(t *testing.T) {
	const autrec = `
	<marc:record format="MARC21" type="Authority" id="90067942" xmlns:marc="info:lc/xmlns/marcxchange-v1">
		<marc:leader>99999nz  a2299999n  4500</marc:leader>
		<marc:controlfield tag="001">90067942</marc:controlfield>
		<marc:controlfield tag="003">NO-TrBIB</marc:controlfield>
		<marc:controlfield tag="005">20220118132609.0</marc:controlfield>
		<marc:controlfield tag="008">120319n| adz|naabn|         |a|ana|     </marc:controlfield>
		<marc:datafield tag="024" ind1="7" ind2=" ">
			<marc:subfield code="a">x90067942</marc:subfield>
			<marc:subfield code="2">NO-TrBIB</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="024" ind1="7" ind2=" ">
			<marc:subfield code="a">http://hdl.handle.net/11250/470737</marc:subfield>
			<marc:subfield code="2">hdl</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="024" ind1="7" ind2=" ">
			<marc:subfield code="a">0000000109115902</marc:subfield>
			<marc:subfield code="2">isni</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="024" ind1="7" ind2=" ">
			<marc:subfield code="a">http://viaf.org/viaf/59247880</marc:subfield>
			<marc:subfield code="2">viaf</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="024" ind1="7" ind2=" ">
			<marc:subfield code="a">https://id.bs.no/bibbi/37524</marc:subfield>
			<marc:subfield code="2">bibbi</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="040" ind1=" " ind2=" ">
			<marc:subfield code="a">NO-TrBIB</marc:subfield>
			<marc:subfield code="b">nob</marc:subfield>
			<marc:subfield code="c">NO-TrBIB</marc:subfield>
			<marc:subfield code="f">noraf</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="043" ind1=" " ind2=" ">
			<marc:subfield code="c">fi</marc:subfield>
			<marc:subfield code="c">no</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="100" ind1="1" ind2=" ">
			<marc:subfield code="a">Valkeapää, Nils-Aslak</marc:subfield>
			<marc:subfield code="d">1943-2001</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="370" ind1=" " ind2=" ">
			<marc:subfield code="a">Palojoensuu, Enontekis, Lappland, Finland</marc:subfield>
			<marc:subfield code="b">Espo, Finland</marc:subfield>
			<marc:subfield code="c">Skibotn, Storfjord, Troms, Norge</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="375" ind1=" " ind2=" ">
			<marc:subfield code="a">m</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="386" ind1=" " ind2=" ">
			<marc:subfield code="a">n.-sam.</marc:subfield>
			<marc:subfield code="m">Nasjonalitet/regional gruppe</marc:subfield>
			<marc:subfield code="2">bs-nasj</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="400" ind1="0" ind2=" ">
			<marc:subfield code="a">Áillohaš</marc:subfield>
			<marc:subfield code="d">1943-2001</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="400" ind1="0" ind2=" ">
			<marc:subfield code="a">Áilu</marc:subfield>
			<marc:subfield code="d">1943-2001</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="901" ind1=" " ind2=" ">
			<marc:subfield code="a">kat3</marc:subfield>
		</marc:datafield>
	</marc:record>
	`

	want := sirkulator.Resource{
		Type:  sirkulator.TypePerson,
		Label: "Nils-Aslak Valkeapää (1943–2001)",
		Data: sirkulator.Person{
			Name:           "Nils-Aslak Valkeapää",
			NameVariations: []string{"Áillohaš", "Áilu"},
			YearRange: sirkulator.YearRange{
				From: 1943,
				To:   2001,
			},
			Gender:            vocab.GenderMale,
			PlaceAssociations: []string{"marc/fi", "marc/no", "bs/n.", "bs/sam."},
		},
		Links: [][2]string{{"bibsys", "90067942"}, {"isni", "0000000109115902"}, {"viaf", "59247880"}, {"bibbi", "37524"}},
	}

	got, err := PersonFromAuthority(marc.MustParseString(autrec))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("PersonFromAuthority() mismatch (-want +got):\n%s", diff)
	}
}

func TestCorporationFromAuthority(t *testing.T) {
	const autrec = `
	<marc:record format="MARC21" type="Authority" id="11008994" xmlns:marc="info:lc/xmlns/marcxchange-v1">
		<marc:leader>99999nz  a2299999n  4500</marc:leader>
		<marc:controlfield tag="001">11008994</marc:controlfield>
		<marc:controlfield tag="003">NO-TrBIB</marc:controlfield>
		<marc:controlfield tag="005">20180525124711.0</marc:controlfield>
		<marc:controlfield tag="008">110113n| adz|naabn|         |a|ana|     </marc:controlfield>
		<marc:datafield tag="024" ind1="7" ind2=" ">
			<marc:subfield code="a">x11008994</marc:subfield>
			<marc:subfield code="2">NO-TrBIB</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="024" ind1="7" ind2=" ">
			<marc:subfield code="a">http://hdl.handle.net/11250/834336</marc:subfield>
			<marc:subfield code="2">hdl</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="040" ind1=" " ind2=" ">
			<marc:subfield code="a">NO-TrBIB</marc:subfield>
			<marc:subfield code="b">nob</marc:subfield>
			<marc:subfield code="c">NO-TrBIB</marc:subfield>
			<marc:subfield code="f">noraf</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="110" ind1="2" ind2=" ">
			<marc:subfield code="a">United Nations </marc:subfield>
			<marc:subfield code="b">Working Group on Indigenous Populations </marc:subfield>
		</marc:datafield>
		<marc:datafield tag="410" ind1="2" ind2=" ">
			<marc:subfield code="a">Forente nasjoner </marc:subfield>
			<marc:subfield code="b">Arbeidsgruppen for urfolksspørsmål</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="410" ind1="2" ind2=" ">
			<marc:subfield code="a">Arbeidsgruppen for urfolksspørsmål</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="410" ind1="2" ind2=" ">
			<marc:subfield code="a">Working Group on Indigenous Populations</marc:subfield>
		</marc:datafield>
		<marc:datafield tag="901" ind1=" " ind2=" ">
			<marc:subfield code="a">kat3</marc:subfield>
		</marc:datafield>
	</marc:record>
	`

	want := sirkulator.Resource{
		Type:  sirkulator.TypeCorporation,
		Label: "Working Group on Indigenous Populations / United Nations",
		Data: sirkulator.Corporation{
			Name:           "Working Group on Indigenous Populations",
			ParentName:     "United Nations",
			NameVariations: []string{"Arbeidsgruppen for urfolksspørsmål / Forente nasjoner", "Arbeidsgruppen for urfolksspørsmål"},
		},
		Links: [][2]string{{"bibsys", "11008994"}},
	}

	got, err := CorporationFromAuthority(marc.MustParseString(autrec))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CorporationFromAuthority() mismatch (-want +got):\n%s", diff)
	}
}
