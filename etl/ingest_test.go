package etl

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/google/go-cmp/cmp"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/sql"
)

func mustGzip(s string) []byte {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(s)); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func mustJson(o interface{}) []byte {
	b, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return b
}

const isbn8213002962 = `
<record xmlns="http://www.loc.gov/MARC21/slim">
    <leader>01659cam a2200457 c 4500</leader>
    <controlfield tag="001">998110670684702201</controlfield>
    <controlfield tag="005">20211030210604.0</controlfield>
    <controlfield tag="007">ta</controlfield>
    <controlfield tag="007">cr||||||||||||</controlfield>
    <controlfield tag="008">150827s1980    no#||||| |||||000|0|nob|d</controlfield>
    <datafield ind1=" " ind2=" " tag="015">
        <subfield code="a">8001799</subfield>
        <subfield code="2">nbf</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="020">
        <subfield code="a">8202018560</subfield>
        <subfield code="q">h.</subfield>
        <subfield code="c">Nkr 73.00</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="035">
        <subfield code="a">811067068-47bibsys_network</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="035">
        <subfield code="a">(NO-TrBIB)811067068</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="035">
        <subfield code="a">(NO-TrBIB)140695214</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="040">
        <subfield code="a">NO-TrBIB</subfield>
        <subfield code="b">nob</subfield>
        <subfield code="e">katreg</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="044">
        <subfield code="c">no</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="080">
        <subfield code="a">582.26</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="080">
        <subfield code="a">582.26 (084.11)</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="080">
        <subfield code="a">582.26(481)</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="080">
        <subfield code="a">582.26:581.9</subfield>
    </datafield>
    <datafield ind1="7" ind2="4" tag="082">
        <subfield code="a">589.3</subfield>
        <subfield code="q">NO-OsNB</subfield>
        <subfield code="2">3/nor</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="084">
        <subfield code="a">zmnp BFM220</subfield>
        <subfield code="2">oosk</subfield>
    </datafield>
    <datafield ind1="1" ind2=" " tag="100">
        <subfield code="a">Åsen, Per Arvid</subfield>
        <subfield code="d">1949-</subfield>
        <subfield code="0">(NO-TrBIB)90294124</subfield>
    </datafield>
    <datafield ind1="1" ind2="0" tag="245">
        <subfield code="a">Illustrert algeflora</subfield>
        <subfield code="c">Per Arvid Åsen ; [illustrasjoner, Per Arvid Åsen]</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="260">
        <subfield code="a">[Oslo]</subfield>
        <subfield code="b">Cappelen</subfield>
        <subfield code="c">1980</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="300">
        <subfield code="a">63 s.</subfield>
        <subfield code="b">ill.</subfield>
        <subfield code="c">4°</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="533">
        <subfield code="a">Elektronisk reproduksjon</subfield>
        <subfield code="b">[Norge]</subfield>
        <subfield code="c">Nasjonalbiblioteket Digital</subfield>
        <subfield code="d">2014-03-08</subfield>
    </datafield>
    <datafield ind1=" " ind2="7" tag="650">
        <subfield code="a">Hav</subfield>
        <subfield code="2">noubomn</subfield>
        <subfield code="0">(NO-TrBIB)REAL000761</subfield>
    </datafield>
    <datafield ind1=" " ind2="7" tag="650">
        <subfield code="a">Botanikk</subfield>
        <subfield code="2">noubomn</subfield>
        <subfield code="0">(NO-TrBIB)REAL013830</subfield>
    </datafield>
    <datafield ind1=" " ind2="7" tag="650">
        <subfield code="a">Alger</subfield>
        <subfield code="2">noubomn</subfield>
        <subfield code="0">(NO-TrBIB)REAL012685</subfield>
    </datafield>
    <datafield ind1=" " ind2="7" tag="650">
        <subfield code="a">Tang og tare</subfield>
        <subfield code="2">tekord</subfield>
    </datafield>
    <datafield ind1=" " ind2="7" tag="650">
        <subfield code="a">Alger</subfield>
        <subfield code="2">tekord</subfield>
    </datafield>
    <datafield ind1=" " ind2="7" tag="651">
        <subfield code="a">Norge</subfield>
        <subfield code="2">noubomn</subfield>
        <subfield code="0">(NO-TrBIB)REAL030753</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="653">
        <subfield code="a">alger</subfield>
        <subfield code="a">botanikk</subfield>
        <subfield code="a">havet</subfield>
        <subfield code="a">norge</subfield>
        <subfield code="a">floraer</subfield>
        <subfield code="a">bestemmelseslitteratur</subfield>
    </datafield>
    <datafield ind1=" " ind2="7" tag="655">
        <subfield code="a">Floraer</subfield>
        <subfield code="2">noubomn</subfield>
        <subfield code="0">(NO-TrBIB)REAL030119</subfield>
    </datafield>
    <datafield ind1=" " ind2="7" tag="655">
        <subfield code="a">Bestemmelseslitteratur</subfield>
        <subfield code="2">noubomn</subfield>
        <subfield code="0">(NO-TrBIB)REAL030073</subfield>
    </datafield>
    <datafield ind1="1" ind2=" " tag="700">
        <subfield code="a">Åsen, Per Arvid</subfield>
        <subfield code="d">1949-</subfield>
        <subfield code="4">ill</subfield>
        <subfield code="0">(NO-TrBIB)90294124</subfield>
    </datafield>
    <datafield ind1="4" ind2="1" tag="856">
        <subfield code="3">Fulltekst</subfield>
        <subfield code="u">https://www.nb.no/search?q=oaiid:"oai:nb.bibsys.no:998110670684702202"&amp;mediatype=bøker</subfield>
        <subfield code="y">Nettbiblioteket</subfield>
        <subfield code="z">Søke-URL</subfield>
    </datafield>
    <datafield ind1="4" ind2="2" tag="856">
		<subfield code="3">Omslagsbilde</subfield>
		<subfield code="u">%[1]s/invalid.jpg</subfield>
		<subfield code="q">image/jpeg</subfield>
	</datafield>
    <datafield ind1="4" ind2="2" tag="856">
        <subfield code="3">Originalt bilde</subfield>
        <subfield code="u">%[1]s/img.jpg</subfield>
        <subfield code="q">image/jpeg</subfield>
    </datafield>
    <datafield ind1="4" ind2="2" tag="856">
        <subfield code="3">Miniatyrbilde</subfield>
        <subfield code="u">https://contents.bibs.aws.unit.no/files/images/small/7/6/9788213002967.jpg</subfield>
        <subfield code="q">image/jpeg</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="901">
        <subfield code="a">80</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="913">
        <subfield code="a">Norbok</subfield>
        <subfield code="b">NB</subfield>
    </datafield>
</record>
`

func TestIngestISBN(t *testing.T) {
	// Setup mock server for testing that image is downloaded
	// TODO actually test this when Ingest downloads..
	const validImagePath = "/img.jpg"
	const invalidImagePath = "/invalid.jpg"
	validJpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xC2, 0x00, 0x0B, 0x08, 0x00, 0x01, 0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xFF, 0xC4, 0x00, 0x14, 0x10, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01, 0x00, 0x01, 0x3F, 0x10}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case validImagePath:
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(validJpeg)
		case invalidImagePath:
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte{1, 2, 3})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	marcxml := fmt.Sprintf(isbn8213002962, ts.URL)
	oairecord := mustGzip(marcxml)

	// Setup and populate DB
	db, err := sql.OpenMem()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	conn := db.Get(nil)
	defer db.Put(conn)

	// Insert person resource in DB
	personWant := sirkulator.Person{
		Name: "Per Arvid Åsen",
		YearRange: sirkulator.YearRange{
			FromYear: 1949,
		},
	}
	b := mustJson(personWant)

	q := fmt.Sprintf(`
			INSERT INTO oai.source (id, url, dataset, prefix)
				VALUES ('bibsys/pub','dummy','dummy','dummy');
			INSERT INTO oai.record (source_id, id, data, created_at, updated_at)
				VALUES ('bibsys/pub', '999608854204702201', x'%x', 0, 0);
			INSERT INTO oai.record_id (source_id, record_id, type, id)
				VALUES ('bibsys/pub', '999608854204702201', 'isbn', '8213002962');
			INSERT INTO resource (id, type, label, data, created_at, updated_at)
				VALUES ('p0','person', 'Per Arvid Åsen (1949-)', x'%x', 0, 0);
			INSERT INTO link (resource_id, type, id)
				VALUES ('p0', 'bibsys', '90294124');
		`, oairecord, b)

	if err := sqlitex.ExecScript(conn, q); err != nil {
		t.Fatal(err)
	}

	// Ingest by ISBN number
	ing := NewIngestor(db)
	ing.idFunc = testID()
	if err := ing.IngestISBN(context.Background(), "8213002962"); err != nil {
		t.Fatal(err)
	}

	// Verify that the existing person resource is still there unchanged
	perWant := sirkulator.Resource{
		Type:  sirkulator.TypePerson,
		Label: "Per Arvid Åsen (1949-)",
		ID:    "p0",
		Data:  &personWant,
	}
	perGot, err := sql.GetResource(conn, sirkulator.TypePerson, "p0")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(perWant, perGot); diff != "" {
		t.Errorf("person mismatch (-want +got):\n%s", diff)
	}

	// Verify that publication got stored
	pubData := sirkulator.Publication{
		Title:      "Illustrert algeflora",
		Language:   "nob",
		Nonfiction: true,
		Year:       1980,
		NumPages:   63,
		Publisher:  "Cappelen",
		GenreForms: []string{"Floraer", "Bestemmelseslitteratur"},
	}
	pubWant := sirkulator.Resource{
		Type:  sirkulator.TypePublication,
		Label: "Per Arvid Åsen - Illustrert algeflora (1980)",
		ID:    "t1",
		Data:  &pubData,
	}
	pubGot, err := sql.GetResource(conn, sirkulator.TypePublication, "t1")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(pubWant, pubGot); diff != "" {
		t.Errorf("publication mismatch (-want +got):\n%s", diff)
	}

	// Verify that 2 relations where added, from existing person to imported publication
	wantRelations := []sirkulator.Relation{
		{
			FromID: "t1",
			ToID:   "p0",
			Type:   "has_contributor",
			Data:   map[string]interface{}{"role": "aut", "main_entry": true},
		},
		{
			FromID: "t1",
			ToID:   "p0",
			Type:   "has_contributor",
			Data:   map[string]interface{}{"role": "ill"},
		},
	}
	var gotRelations []sirkulator.Relation
	checkRel := func(stmt *sqlite.Stmt) error {
		rel := sirkulator.Relation{
			FromID: stmt.ColumnText(0),
			ToID:   stmt.ColumnText(1),
			Type:   stmt.ColumnText(2),
		}
		data := make(map[string]interface{})
		if err := json.Unmarshal([]byte(stmt.ColumnText(3)), &data); err != nil {
			return err
		}
		rel.Data = data
		gotRelations = append(gotRelations, rel)
		return nil
	}
	if err := sqlitex.Exec(conn, "SELECT from_id, to_id, type, data FROM relation", checkRel); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantRelations, gotRelations); diff != "" {
		t.Errorf("relations mismatch (-want +got):\n%s", diff)
	}

	// Verify that reviews where stored
	wantReviews := []sirkulator.Relation{
		{
			FromID: "t1",
			Type:   "published_by",
			Data:   map[string]interface{}{"label": "Cappelen"},
		},
	}
	var gotReviews []sirkulator.Relation
	checkRev := func(stmt *sqlite.Stmt) error {
		rel := sirkulator.Relation{
			FromID: stmt.ColumnText(0),
			Type:   stmt.ColumnText(1),
		}
		data := make(map[string]interface{})
		if err := json.Unmarshal([]byte(stmt.ColumnText(2)), &data); err != nil {
			return err
		}
		rel.Data = data
		gotReviews = append(gotReviews, rel)
		return nil
	}
	const rq = "SELECT from_id, type, data FROM review WHERE from_id=?"
	if err := sqlitex.Exec(conn, rq, checkRev, "t1"); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantReviews, gotReviews); diff != "" {
		t.Errorf("relations mismatch (-want +got):\n%s", diff)
	}

	// TODO link to nb deduced from marc record? /nb, nb-free, nb-norway?
	// TODO check for link [2]string{"nb-fulltext","2014021108035"}
	// https://api.nb.no/catalog/v1/items?q=oaiid%3A%22oai%3Anb.bibsys.no%3A998110670684702202%22&filter=mediatype%3Aaviser%20OR%20mediatype%3Abilder%20OR%20mediatype%3Ab%C3%B8ker%20OR%20mediatype%3Akart%20OR%20mediatype%3Amusikk%20OR%20mediatype%3Amusikkmanuskripter%20OR%20mediatype%3Anoter%20OR%20mediatype%3Aplakater%20OR%20mediatype%3Aprivatarkivmateriale%20OR%20mediatype%3Atidsskrift&aggs=mediatype&size=1&profile=nbdigital
	// https://urn.nb.no/URN:NBN:no-nb_digibok_2014021108035
	// https://www.nb.no/items/4b5337744e197a56fa0aeb2df01feb60?page=0
}
