package dewey

import (
	"compress/bzip2"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/rdf"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/client"
	"github.com/knakk/sirkulator/search"
	"github.com/knakk/sirkulator/vocab"
	"github.com/knakk/sparql"
)

type importDewey struct {
	Number     string
	Notaion    string
	Parent     string
	Name       string
	Terms      []string
	Parts      []string
	Deprecated bool
	Created    string // YYYY-mm-dd
	Updated    string // YYYY-mm-dd
}

type importBatch struct {
	Resources []sirkulator.Resource
	Relations []sirkulator.Relation
}

type FusekiDescribeNT struct {
	Query string
}

func (f FusekiDescribeNT) GenRequest(endpoint string) (*http.Request, error) {
	url := fmt.Sprintf("%s?query=%s&format=nt", endpoint, url.QueryEscape(f.Query))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	return req, err
}

type ImportJob struct {
	wg        sync.WaitGroup // keeping track of indexing TODO consider sync.ErrGroup
	peek      *rdf.Triple
	peekParts []string

	// Setup:
	DB        *sqlitex.Pool
	Idx       *search.Index
	BatchSize int // Number of resources to persist to DB at a time
	Update    bool
}

func (j *ImportJob) Name() string {
	if j.Update {
		return "dewey_import_update"
	}
	return "dewey_import_all"
}

func (j *ImportJob) open(ctx context.Context, w io.Writer) (io.ReadCloser, error) {
	if j.Update {
		conn := j.DB.Get(ctx)
		if conn == nil {
			return nil, context.Canceled
		}
		stmt := conn.Prep(`
			SELECT max(date(datetime(updated_at, 'unixepoch'))) AS latest
			  FROM resource
			 WHERE type='dewey'`)

		date, err := sqlitex.ResultText(stmt)
		if err != nil {
			return nil, err
		}
		sparqlEndpoint := "https://data.ub.uio.no/sparql"
		fmt.Fprintf(w, "finding most recent dewey update in DB: %s\n", date)
		fmt.Fprintf(w, "querying sparql endpoint at %s for updates since %s\n", sparqlEndpoint, date)
		repo, err := sparql.NewRepo(sparqlEndpoint)
		if err != nil {
			return nil, err
		}
		q := fmt.Sprintf(`
		PREFIX skos: <http://www.w3.org/2004/02/skos/core#>
		PREFIX purl: <http://purl.org/dc/terms/>
		PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

		DESCRIBE ?d WHERE {
		?d a skos:Concept ;
			skos:inScheme <http://dewey.info/scheme/edition/e23/> ;
			purl:created ?created ;
			purl:modified ?modified .

		FILTER (?modified >= "%s"^^xsd:date)
		}`, date)
		f, err := repo.QueryWithoutParsing(FusekiDescribeNT{Query: q})
		if err != nil {
			return nil, err
		}
		return f, nil
	}
	fmt.Fprintln(w, "downloading webdewey n-triples dump from url: https://data.ub.uio.no/dumps/wdno.nt.bz2")
	f, err := client.Open(ctx, "https://data.ub.uio.no/dumps/wdno.nt.bz2")
	return f, err
}

func (j *ImportJob) Run(ctx context.Context, w io.Writer) error {
	f, err := j.open(ctx, w)
	if err != nil {
		return err
	}
	defer f.Close()

	var d rdf.TripleDecoder
	if j.Update {
		d = rdf.NewTripleDecoder(f, rdf.NTriples)
	} else {
		d = rdf.NewTripleDecoder(bzip2.NewReader(f), rdf.NTriples)
	}

	fmt.Fprintln(w, "start importing deweynumbers")
	c := 0
	for {
		batch, err := j.getBatch(d)
		if err != nil {
			return err
		}
		if len(batch.Resources) == 0 {
			break
		}

		if err := j.persist(ctx, batch); err != nil {
			return err
		}
		c += len(batch.Resources)

		go j.index(batch.Resources)
		w.Write([]byte("."))
	}
	w.Write([]byte("\n"))

	conn := j.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer j.DB.Put(conn)

	if err := sqlitex.ExecScript(conn, `
		UPDATE relation
		SET
			to_id = candidates.new_to_id,
			data = json_remove(data, '$.to_id'),
			queued_at = NULL
		FROM (
			SELECT rel.id, rel.from_id, json_extract(rel.data ,'$.to_id') AS new_to_id
			  FROM relation rel
			  JOIN resource res ON (json_extract(rel.data ,'$.to_id')=res.id)
			 WHERE rel.to_id IS NULL)
			AS candidates
		WHERE relation.id=candidates.id;`); err != nil {
		return err
	}

	// Some duplicate relations might be inserted, delete them at this point:
	if err := sqlitex.ExecScript(conn,
		`DELETE FROM relation WHERE to_id IS NOT NULL AND rowid NOT IN
		 (SELECT min(rowid) FROM relation GROUP BY from_id, to_id, type);
		 DELETE FROM relation WHERE to_id IS NULL AND rowid NOT IN
		 (SELECT min(rowid) FROM relation GROUP BY from_id, data, type)`); err != nil {
		fmt.Fprintf(w, "failed to delete duplicate relations: %v\n", err)
	}

	// Wait for indexing to complete
	j.wg.Wait()

	if j.Update {
		fmt.Fprintf(w, "done updating/importing %d dewey numbers\n", c)
		// TODO list numbers if less than < 30?
	} else {
		fmt.Fprintf(w, "done importing %d dewey numbers\n", c)
	}

	return nil
}

var rxpAddNumber = regexp.MustCompile(`\d(A|B|C)?--\d{1,3}\.\d+`)

func appendIfNew(existing []string, val string) []string {
	for _, v := range existing {
		if v == val {
			return existing
		}
	}
	existing = append(existing, val)
	return existing
}

func (j *ImportJob) getBatch(d rdf.TripleDecoder) (importBatch, error) {

	numbers := make(map[string]importDewey, j.BatchSize)

	var (
		subj  string
		dewey importDewey
		tr    rdf.Triple
		err   error
		parts []string
	)
	parts = append(parts, j.peekParts...)
	j.peekParts = j.peekParts[:0]

	for {
		if j.peek != nil {
			tr = *j.peek
			j.peek = nil
		} else {
			tr, err = d.Decode()
			if err == io.EOF {
				break
			} else if err != nil {
				return importBatch{}, err
			}
		}
		if s := tr.Subj.String(); strings.HasPrefix(s, "http://dewey.info/class/") {
			s = strings.TrimPrefix(s, "http://dewey.info/class/")
			s = strings.TrimSuffix(s, "/e23/")
			if strings.Contains(s, "--") {
				s = "T" + s
			}
			if s != subj {
				if len(numbers) == j.BatchSize {
					j.peek = &tr
					j.peekParts = parts
					break
				}
				if subj != "" {
					numbers[subj] = dewey
				}
				subj = s
				dewey = numbers[subj]
				dewey.Number = subj
				dewey.Parts = append(dewey.Parts, parts...)
				parts = parts[:0]
			}
		}

		switch tr.Pred.String() {
		case "http://www.w3.org/2004/02/skos/core#prefLabel":
			dewey.Name = strings.TrimSuffix(tr.Obj.String(), "@nb")
		case "http://www.w3.org/2004/02/skos/core#altLabel":
			dewey.Terms = appendIfNew(dewey.Terms, strings.TrimSuffix(tr.Obj.String(), "@nb"))
		case "http://purl.org/dc/terms/created":
			if l, ok := tr.Obj.(rdf.Literal); ok {
				dewey.Created = l.String()
			}
		case "http://purl.org/dc/terms/modified":
			if l, ok := tr.Obj.(rdf.Literal); ok {
				dewey.Updated = l.String()
			}
		case "http://www.w3.org/1999/02/22-rdf-syntax-ns#first":
			// Blank nodes are always preceding the dewey number which points to them, so
			// This belongs to dewey "next" number in the file.
			// It's a tad brittle to base the parsing on the order of the triples..
			// TODO consider more robust parsing
			if o := tr.Obj.String(); strings.HasPrefix(o, "http://dewey.info/class/") {
				o = strings.TrimPrefix(o, "http://dewey.info/class/")
				o = strings.TrimSuffix(o, "/e23/")
				if rxpAddNumber.MatchString(o) {
					// convert eg '6--919.94' to '6--91994'
					o = strings.Replace(o, ".", "", -1)
				}
				if strings.Contains(o, "--") {
					o = "T" + o
				}
				parts = appendIfNew(parts, o)
			}
		case "http://www.w3.org/2004/02/skos/core#broader":
			if o := tr.Obj.String(); strings.HasPrefix(o, "http://dewey.info/class/") {
				o = strings.TrimPrefix(o, "http://dewey.info/class/")
				o = strings.TrimSuffix(o, "/e23/")
				if strings.Contains(o, "--") {
					o = "T" + o
				}
				dewey.Parent = o
			}
		case "http://www.w3.org/2004/02/skos/core#notation":
			dewey.Notaion = tr.Obj.String()
		case "http://www.w3.org/2002/07/owl#deprecated":
			dewey.Deprecated = true
		}
	}
	if dewey.Number != "" {
		numbers[subj] = dewey
	}

	batch := importBatch{
		Resources: make([]sirkulator.Resource, 0, len(numbers)),
	}

	for _, n := range numbers {
		data := sirkulator.Dewey{
			Number: n.Number,
			Name:   n.Name,
			Terms:  n.Terms,
		}
		res := sirkulator.Resource{
			ID:    n.Number,
			Type:  sirkulator.TypeDewey,
			Label: data.Label(),
			Data:  data,
		}
		if created, err := time.Parse("2006-01-02", n.Created); err == nil {
			res.CreatedAt = created
		}
		if updated, err := time.Parse("2006-01-02", n.Updated); err == nil {
			res.UpdatedAt = updated
		}
		if n.Deprecated {
			res.ArchivedAt = res.UpdatedAt
		}
		batch.Resources = append(batch.Resources, res)

		if n.Parent != "" {
			rel := sirkulator.Relation{
				FromID: n.Number,
				Type:   string(vocab.RelationHasParent),
				ToID:   n.Parent,
			}
			batch.Relations = append(batch.Relations, rel)
		}

		for _, part := range n.Parts {
			rel := sirkulator.Relation{
				FromID: n.Number,
				Type:   string(vocab.RelationHasPart),
				ToID:   part,
			}
			batch.Relations = append(batch.Relations, rel)
		}

	}

	return batch, nil
}

func (j *ImportJob) persist(ctx context.Context, batch importBatch) (err error) {
	conn := j.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer j.DB.Put(conn)
	defer sqlitex.Save(conn)(&err)

	for _, res := range batch.Resources {
		stmt := conn.Prep(`
			INSERT INTO resource (type, id, label, data, created_at, updated_at, archived_at)
				VALUES ($type, $id, $label, $data, $created, $updated, $archived)
			ON CONFLICT(id) DO UPDATE
				SET label=$label,
					data=$data,
					updated_at=$updated,
					archived_at=$archived
		`)

		stmt.SetText("$type", res.Type.String())
		stmt.SetText("$id", res.ID)
		stmt.SetText("$label", res.Label)
		stmt.SetInt64("$created", res.CreatedAt.Unix())
		stmt.SetInt64("$updated", res.UpdatedAt.Unix())
		if !res.ArchivedAt.IsZero() {
			stmt.SetInt64("$archived", res.ArchivedAt.Unix())
		} else {
			stmt.SetNull("$archived")
		}

		b, err := json.Marshal(res.Data)
		if err != nil {
			return err // TODO annotate
		}
		stmt.SetBytes("$data", b)
		if _, err := stmt.Step(); err != nil {
			return err // TODO annotate
		}
	}

	for _, rel := range batch.Relations {
		delQ := `DELETE FROM relation WHERE id=? AND (type='has_parent' OR type='has_part')`
		if err := sqlitex.Exec(conn, delQ, nil, rel.FromID); err != nil {
			return err // TODO annotate
		}
		stmt := conn.Prep(`
			WITH v(from_id, type, data) AS (VALUES ($from_id, $type, $data))
			INSERT INTO relation (from_id, to_id, type, data, queued_at)
			SELECT
				v.from_id,
				res.id,
				v.type,
				json_patch(v.data, iif(res.id IS NULL,
					iif($to_id != '', json_object('to_id', $to_id), '{}'), '{}')) AS data,
				iif(res.id IS NULL, $queued_at, NULL) AS queued_at
			FROM v LEFT JOIN resource res ON (res.id = $to_id)
		`)

		stmt.SetText("$from_id", rel.FromID)
		stmt.SetText("$to_id", rel.ToID)
		stmt.SetText("$type", rel.Type)
		stmt.SetInt64("$queued_at", time.Now().Unix())
		b, err := json.Marshal(rel.Data)
		if err != nil {
			return err // TODO annotate
		}
		stmt.SetBytes("$data", b)
		if _, err := stmt.Step(); err != nil {
			return err // TODO annotate
		}
	}

	return nil
}

func (j *ImportJob) index(batch []sirkulator.Resource) {
	j.wg.Add(1)
	defer j.wg.Done()

	docs := make([]search.Document, 0, len(batch))
	for _, r := range batch {
		docs = append(docs, r.Document())
	}
	if err := j.Idx.Store(docs...); err != nil {
		log.Printf("ImportJob.index: %v", err) // TODO remove, or write to w
	}
}
