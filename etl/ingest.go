// package etl
package etl

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/marc"
	"github.com/knakk/sirkulator/oai"
)

type Ingestor struct {
	db     *sqlitex.Pool
	idFunc func() string

	// Options
	ImageWidth int
	ImageWebp  bool
}

func NewIngestor(db *sqlitex.Pool) *Ingestor {
	return &Ingestor{
		db:         db,
		ImageWidth: 200,                 // default resize width
		idFunc:     sirkulator.GetNewID, // can be overwritten with a deterministic function in tests
	}
}

func (ig *Ingestor) IngestISBN(ctx context.Context, isbn string) error {
	// First, check if publication is allready in DB
	// TODO
	//
	// TODO future idea: maybe if in local DB, go look on the internet and see if
	//      we can gain something new info diffing existing record,
	//      but then, we have to consider the changes we have made (in resource_edit_log?)

	// Next, check if record is available in local oai DB
	rec, err := ig.localRecord(ctx, "isbn", isbn)
	if errors.Is(err, sirkulator.ErrNotFound) {
		// Record not in local DB; go search the internet
		rec, err = ig.remoteRecord("isbn", isbn)
	}
	if err != nil {
		return err
	}

	data, err := ingestMarcRecord(rec.Source, rec.Data, ig.idFunc)
	if err != nil {
		return err
	}
	return ig.Ingest(ctx, data)
}

func (ig *Ingestor) localRecord(ctx context.Context, idtype, id string) (oai.Record, error) {
	var rec oai.Record
	conn := ig.db.Get(ctx)
	if conn == nil {
		return rec, context.Canceled
	}
	defer ig.db.Put(conn)

	fn := func(stmt *sqlite.Stmt) error {
		rec.Source = stmt.ColumnText(0)
		rec.ID = stmt.ColumnText(1)
		blob, err := conn.OpenBlob("oai", "record", "data", stmt.ColumnInt64(2), false)
		if err != nil {
			return err
		}
		defer blob.Close()
		gz, err := gzip.NewReader(blob)
		if err != nil {
			return err
		}
		dec := marc.NewDecoder(gz)
		mrc, err := dec.Decode()
		if err != nil {
			return err
		}
		rec.Data = mrc
		return nil
	}
	q := `
		SELECT
			r.source_id,
			r.id,
			r.rowid
		FROM oai.record_id t
			JOIN oai.record r ON (t.source_id=r.source_id AND t.record_id=r.id)
		WHERE t.type=? AND t.id=?
	`
	if err := sqlitex.Exec(conn, q, fn, idtype, id); err != nil {
		return rec, err
	}
	// TODO what if rec is emtpty? check for rec.ID != ""

	return rec, nil
}

func (ig *Ingestor) remoteRecord(idtype, id string) (oai.Record, error) {
	return oai.Record{}, errors.New("remoteRecord: TODO")
}

// persistIngestion will store all resources, relations and reviews in
// the Ingestion. No further validation of input is performed - all of
// the given data is assumed to be valid at this point, as not to
// trigger any SQL constraints when inserting into DB.
// TODO: handle data.Covers
func persistIngestion(conn *sqlite.Conn, data Ingestion) (err error) {
	defer sqlitex.Save(conn)(&err)

	for _, res := range data.Resources {
		stmt := conn.Prep(`
			INSERT INTO resource (type, id, label, data, created_at, updated_at)
				VALUES ($type, $id, $label, $data, 0, 0)
		`)

		stmt.SetText("$type", res.Type.String())
		stmt.SetText("$id", res.ID)
		stmt.SetText("$label", res.Label)
		b, err := json.Marshal(res.Data)
		if err != nil {
			return err // TODO annotate
		}
		stmt.SetBytes("$data", b)
		if _, err := stmt.Step(); err != nil {
			return err // TODO annotate
		}
	}

	for _, rel := range data.Relations {
		stmt := conn.Prep(`
			INSERT INTO relation (from_id, to_id, type, data)
				VALUES ($from_id, $to_id, $type, $data)
		`)

		stmt.SetText("$from_id", rel.FromID)
		stmt.SetText("$to_id", rel.ToID)
		stmt.SetText("$type", rel.Type)
		b, err := json.Marshal(rel.Data)
		if err != nil {
			return err // TODO annotate
		}
		stmt.SetBytes("$data", b)
		if _, err := stmt.Step(); err != nil {
			return err // TODO annotate
		}
	}

	for _, rev := range data.Reviews {
		stmt := conn.Prep(`
			INSERT INTO review (from_id, type, data, queued_at)
				VALUES ($from_id, $type, $data, $queued_at)
		`)
		stmt.SetText("$from_id", rev.FromID)
		stmt.SetText("$type", rev.Type)
		stmt.SetInt64("$queued_at", time.Now().Unix())
		b, err := json.Marshal(rev.Data)
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

func (ig *Ingestor) Ingest(ctx context.Context, data Ingestion) error {
	conn := ig.db.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer ig.db.Put(conn)

	// Check if there are any resources matching any of our local
	// resources in DB, remove from data.Resources and swap the matching IDs in
	// data.Relations.
	for i := len(data.Resources) - 1; i >= 0; i-- {
		res := data.Resources[i]
		if res.Type == sirkulator.TypePublication {
			// This is already checked not in DB, before we started ingesting (TODO)
			// TODO make this (whole Ingest method) transactional?
			continue
		}
		stmt := conn.Prep("SELECT resource_id FROM link WHERE type=$type AND id=$id")
		for _, link := range res.Links {
			stmt.SetText("$type", link[0])
			stmt.SetText("$id", link[1])
			id, err := sqlitex.ResultText(stmt)
			if err != nil {
				return err // TODO annotate
			}
			if id == "" {
				continue
			}
			// Resource is already in our DB and sholdn't be imported
			data.Resources = append(data.Resources[:i], data.Resources[i+1:]...)
			for j, rel := range data.Relations {
				// swap id with exisiting resource id in relations:
				if rel.FromID == res.ID {
					data.Relations[j].FromID = id
				}
				if rel.ToID == res.ID {
					data.Relations[j].ToID = id
				}
			}
			stmt.Reset()
		}
	}

	// Store all resources and relations in a transaction:
	if err := persistIngestion(conn, data); err != nil {
		return err // TODO annotate
	}

	return nil
}

type Ingestion struct {
	Resources []sirkulator.Resource
	Relations []sirkulator.Relation
	Reviews   []sirkulator.Relation
	Covers    []FileFetch
	//Documents []search.Document
}

type FileFetch struct {
	ResourceID string
	IDPair     [2]string
	URL        string
}
