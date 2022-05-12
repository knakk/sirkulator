package etl

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/google/go-cmp/cmp"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/sql"
	"github.com/knakk/sirkulator/vocab"
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

func mustJson(o any) []byte {
	b, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return b
}

const isbn8202018560 = `
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
    <datafield ind1=" " ind2=" " tag="901">
        <subfield code="a">80</subfield>
    </datafield>
    <datafield ind1=" " ind2=" " tag="913">
        <subfield code="a">Norbok</subfield>
        <subfield code="b">NB</subfield>
    </datafield>
</record>
`

func dummyImage() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	img.Set(2, 3, color.RGBA{255, 0, 0, 255})
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, nil); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func TestIngestISBN(t *testing.T) {
	// Setup mock server for testing that image is downloaded
	// TODO actually test this when Ingest downloads..
	const validImagePath = "/img.jpg"
	const invalidImagePath = "/invalid.jpg"

	validJpeg := dummyImage()
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
	marcxml := fmt.Sprintf(isbn8202018560, ts.URL)
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
			From: "1949",
		},
	}
	b := mustJson(personWant)

	q := fmt.Sprintf(`
			INSERT INTO oai.source (id, url, dataset, prefix)
				VALUES ('bibsys/aut','dummy','dummy','dummy'),
				       ('bibsys/pub','dummy','dummy','dummy');
			INSERT INTO oai.record (source_id, id, data, created_at, updated_at)
				VALUES ('bibsys/pub', '999608854204702201', x'%x', 0, 0);
			INSERT INTO resource (id, type, label, data, created_at, updated_at)
				VALUES ('p0','person', 'Per Arvid Åsen (1949-)', x'%x', 0, 0);
			INSERT INTO oai.link (source_id, record_id, type, id)
				VALUES ('bibsys/pub', '999608854204702201', 'isbn', '8202018560');
			INSERT INTO link (resource_id, type, id)
				VALUES ('p0', 'bibsys/aut', '90294124');
		`, oairecord, b)

	if err := sqlitex.ExecScript(conn, q); err != nil {
		t.Fatal(err)
	}

	// Ingest by ISBN number
	ing := NewIngestor(db, nil)
	ing.idFunc = testID()
	ing.ImageDownload = true
	if entry := ing.IngestISBN(context.Background(), "8202018560", true); entry.Error != "" {
		t.Fatal(entry.Error)
	}

	// Verify that the existing person resource is still there unchanged
	perWant := sirkulator.Resource{
		Type:  sirkulator.TypePerson,
		Label: "Per Arvid Åsen (1949-)",
		ID:    "p0",
		Links: [][2]string{{"bibsys/aut", "90294124"}},
		Data:  &personWant,
	}
	perGot, err := sql.GetResource(conn, sirkulator.TypePerson, "p0")
	if err != nil {
		t.Fatal(err)
	}

	// reset timestamps we don't want to compare
	perGot.CreatedAt = time.Time{}
	perGot.UpdatedAt = time.Time{}

	if diff := cmp.Diff(perWant, perGot); diff != "" {
		t.Errorf("person mismatch (-want +got):\n%s", diff)
	}

	// Verify that publication got stored
	pubData := sirkulator.Publication{
		Title:      "Illustrert algeflora",
		Language:   "iso6393/nob",
		Nonfiction: true,
		Year:       "1980",
		NumPages:   "63",
		Binding:    "paperback",
		GenreForms: []string{"Floraer", "Bestemmelseslitteratur"},
	}
	pubWant := sirkulator.Resource{
		Type:  sirkulator.TypePublication,
		Label: "Per Arvid Åsen - Illustrert algeflora (1980)",
		ID:    "t1",
		Links: [][2]string{
			{"bibsys/pub", "998110670684702201"},
			{"isbn", "82-02-01856-0"},
		},
		Data: &pubData,
	}
	pubGot, err := sql.GetResource(conn, sirkulator.TypePublication, "t1")
	if err != nil {
		t.Fatal(err)
	}

	// reset timestamps we don't want to compare
	pubGot.CreatedAt = time.Time{}
	pubGot.UpdatedAt = time.Time{}

	if diff := cmp.Diff(pubWant, pubGot); diff != "" {
		t.Errorf("publication mismatch (-want +got):\n%s", diff)
	}

	// Verify that 2 relations where added, from existing person to imported publication
	wantRelations := []sirkulator.Relation{
		{
			FromID: "t1",
			Type:   "published_by",
			Data:   map[string]any{"label": "Cappelen"},
		},
		{
			FromID: "t1",
			ToID:   "p0",
			Type:   "has_contributor",
			Data:   map[string]any{"role": "aut", "main_entry": true},
		},
		{
			FromID: "t1",
			Type:   "has_classification",
			Data:   map[string]any{"edition": string("3/nor"), "label": string("589.3")},
		},
		{
			FromID: "t1",
			ToID:   "p0",
			Type:   "has_contributor",
			Data:   map[string]any{"role": "ill"},
		},
	}
	var gotRelations []sirkulator.Relation
	checkRel := func(stmt *sqlite.Stmt) error {
		rel := sirkulator.Relation{
			FromID: stmt.ColumnText(0),
			ToID:   stmt.ColumnText(1),
			Type:   stmt.ColumnText(2),
		}
		data := make(map[string]any)
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

	// Verify that ISBN number was stored as link
	if pid, _ := sqlitex.ResultText(conn.Prep("SELECT resource_id FROM link WHERE type='isbn' and id='8202018560'")); pid != "t1" {
		t.Errorf("excepted isbn link from 't1' to 8202018560; got %q", pid)
	}

	// Verify image was downloaded and scaled
	rowid, err := sqlitex.ResultInt64(conn.Prep("SELECT rowid FROM files.image WHERE id='t1';"))
	if err != nil {
		t.Fatal(err)
	}
	blob, err := conn.OpenBlob("files", "image", "data", rowid, false)
	if err != nil {
		t.Fatal(err)
	}
	defer blob.Close()
	img, _, err := image.Decode(blob)
	if err != nil {
		t.Fatal(err)
	}
	if x := img.Bounds().Max.X; x != ing.ImageWidth {
		t.Errorf("image width=%d; expected %d", x, ing.ImageWidth)
	}
	if y := img.Bounds().Max.Y; y != ing.ImageWidth/2 {
		t.Errorf("image width=%d; expected %d", y, ing.ImageWidth/2)
	}
}

const bibsys90294124 = `
<marc:record format="MARC21" type="Authority" id="90294124" xmlns:marc="info:lc/xmlns/marcxchange-v1">
    <marc:leader>99999nz  a2299999n  4500</marc:leader>
    <marc:controlfield tag="001">90294124</marc:controlfield>
    <marc:controlfield tag="003">NO-TrBIB</marc:controlfield>
    <marc:controlfield tag="005">20220118133303.0</marc:controlfield>
    <marc:controlfield tag="008">110315n| adz|naabn|         |a|ana|     </marc:controlfield>
    <marc:datafield tag="024" ind1="7" ind2=" ">
        <marc:subfield code="a">x90294124</marc:subfield>
        <marc:subfield code="2">NO-TrBIB</marc:subfield>
    </marc:datafield>
    <marc:datafield tag="024" ind1="7" ind2=" ">
        <marc:subfield code="a">http://hdl.handle.net/11250/872392</marc:subfield>
        <marc:subfield code="2">hdl</marc:subfield>
    </marc:datafield>
    <marc:datafield tag="024" ind1="7" ind2=" ">
        <marc:subfield code="a">0000000383686038</marc:subfield>
        <marc:subfield code="2">isni</marc:subfield>
    </marc:datafield>
    <marc:datafield tag="024" ind1="7" ind2=" ">
        <marc:subfield code="a">http://viaf.org/viaf/270825492</marc:subfield>
        <marc:subfield code="2">viaf</marc:subfield>
    </marc:datafield>
    <marc:datafield tag="024" ind1="7" ind2=" ">
        <marc:subfield code="a">https://id.bs.no/bibbi/40502</marc:subfield>
        <marc:subfield code="2">bibbi</marc:subfield>
    </marc:datafield>
    <marc:datafield tag="040" ind1=" " ind2=" ">
        <marc:subfield code="a">NO-TrBIB</marc:subfield>
        <marc:subfield code="b">nob</marc:subfield>
        <marc:subfield code="c">NO-TrBIB</marc:subfield>
        <marc:subfield code="f">noraf</marc:subfield>
    </marc:datafield>
    <marc:datafield tag="043" ind1=" " ind2=" ">
        <marc:subfield code="c">no</marc:subfield>
    </marc:datafield>
    <marc:datafield tag="100" ind1="1" ind2=" ">
        <marc:subfield code="a">Åsen, Per Arvid</marc:subfield>
        <marc:subfield code="d">1949-</marc:subfield>
    </marc:datafield>
    <marc:datafield tag="375" ind1=" " ind2=" ">
        <marc:subfield code="a">m</marc:subfield>
    </marc:datafield>
    <marc:datafield tag="386" ind1=" " ind2=" ">
        <marc:subfield code="a">n.</marc:subfield>
        <marc:subfield code="m">Nasjonalitet/regional gruppe</marc:subfield>
        <marc:subfield code="2">bs-nasj</marc:subfield>
    </marc:datafield>
    <marc:datafield tag="901" ind1=" " ind2=" ">
        <marc:subfield code="a">kat3</marc:subfield>
    </marc:datafield>
</marc:record>`

func TestIngestPersonFromLocalOAI(t *testing.T) {
	pubrecord := mustGzip(isbn8202018560)
	autrecord := mustGzip(bibsys90294124)

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

	q := fmt.Sprintf(`
			INSERT INTO oai.source (id, url, dataset, prefix)
				VALUES ('bibsys/aut','dummy','dummy','dummy'),
				       ('bibsys/pub','dummy','dummy','dummy');
			INSERT INTO oai.record (source_id, id, data, created_at, updated_at)
				VALUES ('bibsys/aut', '90294124', x'%x', 0, 0);
			INSERT INTO oai.record (source_id, id, data, created_at, updated_at)
				VALUES ('bibsys/pub', '999608854204702201', x'%x', 0, 0);
			INSERT INTO oai.link (source_id, record_id, type, id)
				VALUES ('bibsys/pub', '999608854204702201', 'isbn', '8202018560');
		`, autrecord, pubrecord)

	if err := sqlitex.ExecScript(conn, q); err != nil {
		t.Fatal(err)
	}

	// Ingest by ISBN number
	ing := NewIngestor(db, nil)
	ing.idFunc = testID()
	if entry := ing.IngestISBN(context.Background(), "8202018560", true); entry.Error != "" {
		t.Fatal(entry.Error)
	}

	// Verify that resource was stored from local authority oai record
	perWant := sirkulator.Resource{
		Type:  sirkulator.TypePerson,
		Label: "Per Arvid Åsen (1949–)",
		ID:    "t2",
		Links: [][2]string{
			{"bibbi", "40502"},
			{"bibsys/aut", "90294124"},
			{"isni", "0000000383686038"},
			{"viaf", "270825492"},
		},
		Data: &sirkulator.Person{
			Name: "Per Arvid Åsen",
			YearRange: sirkulator.YearRange{
				From: "1949",
			},
			Gender:        vocab.GenderMale,
			Countries:     []string{"iso3166/NO"},
			Nationalities: []string{"bs/n"},
		},
	}
	perGot, err := sql.GetResource(conn, sirkulator.TypePerson, "t2")
	if err != nil {
		t.Fatal(err)
	}

	// reset timestamps we don't want to compare
	perGot.CreatedAt = time.Time{}
	perGot.UpdatedAt = time.Time{}

	if diff := cmp.Diff(perWant, perGot); diff != "" {
		t.Errorf("person mismatch (-want +got):\n%s", diff)
	}

}

func TestIngestRemote(t *testing.T) {
	t.Skip("depends on external resource")
	db, err := sql.OpenMem()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	ing := NewIngestor(db, nil)
	ing.idFunc = testID()
	ing.UseRemote = true
	if entry := ing.IngestISBN(context.Background(), "9788253043203", true); entry.Error != "" {
		t.Fatal(entry.Error)
	}
	conn := db.Get(nil)
	defer db.Put(conn)
	fn := func(stmt *sqlite.Stmt) error {
		t.Logf("%s\t%s\t%q\n", stmt.ColumnText(0), stmt.ColumnText(1), stmt.ColumnText(2))
		return nil
	}
	const q = "SELECT id, type, label FROM resource"
	if err := sqlitex.Exec(conn, q, fn); err != nil {
		t.Fatal(err)
	}
	t.FailNow()
}
