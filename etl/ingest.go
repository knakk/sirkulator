// package etl
package etl

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"strings"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/client"
	"github.com/knakk/sirkulator/marc"
	"github.com/knakk/sirkulator/oai"
	"github.com/knakk/sirkulator/search"
	"golang.org/x/image/draw"
)

type Ingestor struct {
	db     *sqlitex.Pool
	idx    *search.Index
	idFunc func() string

	// Options
	UseRemote     bool // if true, use external sources in additin to local
	ImageDownload bool // if true, download images found in imported records
	ImageAsync    bool // if true, download images after IngestISBN has returned
	ImageWidth    int  // scale to this with, calculating width to preserve aspect ratio
	//ImageWebp  bool // convert to webp before storing
}

func NewIngestor(db *sqlitex.Pool, idx *search.Index) *Ingestor {
	return &Ingestor{
		db:         db,
		idx:        idx,
		ImageWidth: 200,                 // default resize width
		idFunc:     sirkulator.GetNewID, // can be overwritten with a deterministic function in tests
	}
}

func NewPreviewIngestor(db *sqlitex.Pool) *Ingestor {
	return &Ingestor{
		db:            db,
		ImageDownload: false,
		idFunc:        func() string { return "" }, // resources with blank IDs will not be persisted
	}
}

var isbnCleaner = strings.NewReplacer("-", "", " ", "") // TODO move to isbn package

func (ig *Ingestor) IngestISBN(ctx context.Context, isbn string) error {
	isbn = isbnCleaner.Replace(isbn)

	// First, check if publication is allready in DB
	found, err := ig.existsISBN(ctx, isbn)
	if err != nil {
		return fmt.Errorf("Ingestor.IngestISBN(%s): %w", isbn, err)
	}
	if found {
		return sirkulator.ErrConflict
	}

	//
	// TODO future idea: maybe if in local DB, go look on the internet and see if
	//      we can gain something new info diffing existing record,
	//      but then, we have to consider the changes we have made (in resource_edit_log?)

	// The preferred source is local oai DB
	var data Ingestion
	rec, err := ig.localRecord(ctx, "isbn", isbn)
	if err == nil {
		data, err = ingestMarcRecord(rec.Source, rec.Data, ig.idFunc)
		if err != nil {
			return err
		}
	} else if errors.Is(err, sirkulator.ErrNotFound) {
		// Record not in local DB - search external sources
		data, err = ig.remoteRecord(ctx, "isbn", isbn)
		if err != nil {
			return err
		}
	}
	// We got data, either from local or remote source, now try to persist it:
	return ig.Ingest(ctx, data)
}

func (ig *Ingestor) existsISBN(ctx context.Context, isbn string) (bool, error) {
	conn := ig.db.Get(ctx)
	if conn == nil {
		return false, context.Canceled
	}
	defer ig.db.Put(conn)
	stmt := conn.Prep("SELECT resource_id FROM link WHERE type='isbn' AND id=$id")
	stmt.SetText("$id", isbn)
	if _, err := sqlitex.ResultText(stmt); err != nil {
		if errors.Is(err, sqlitex.ErrNoResults) {
			return false, nil
		}
		return false, fmt.Errorf("Ingestor.existsISBN(%s): %w ", isbn, err)
	}
	return true, nil
}

func (ig *Ingestor) PreviewISBN(ctx context.Context, isbn string) (Ingestion, error) {
	isbn = isbnCleaner.Replace(isbn)
	var data Ingestion
	rec, err := ig.localRecord(ctx, "isbn", isbn)
	if err == nil {
		data, err = ingestMarcRecord(rec.Source, rec.Data, ig.idFunc)
		if err != nil {
			return data, err
		}
	} else if errors.Is(err, sirkulator.ErrNotFound) {
		data, err = ig.remoteRecord(ctx, "isbn", isbn)
		return data, err
	}
	return data, sirkulator.ErrNotFound
}

// TODO move to sql package?
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
	if rec.ID == "" {
		return rec, sirkulator.ErrNotFound
	}

	return rec, nil
}

type sruResponse struct {
	XMLName         xml.Name `xml:"searchRetrieveResponse"`
	Text            string   `xml:",chardata"`
	Xmlns           string   `xml:"xmlns,attr"`
	Version         string   `xml:"version"`
	NumberOfRecords int      `xml:"numberOfRecords"`
	Records         struct {
		Text   string `xml:",chardata"`
		Record struct {
			Metadata []byte `xml:",innerxml"` // marcxml
		} `xml:"record"`
	} `xml:"records"`
}

// externalSources is a list of prioritized external sources.
var externalSources = []struct {
	Name  string
	Fetch func(ctx context.Context, itdtype, id string, idFunc func() string) (Ingestion, error)
}{
	{
		Name: "bibsys/sru",
		Fetch: func(ctx context.Context, itdtype, id string, idFunc func() string) (Ingestion, error) {
			if itdtype != "isbn" {
				return Ingestion{}, sirkulator.ErrInvalid // ErrUnsupprted?
			}
			const url = "https://bibsys.alma.exlibrisgroup.com/view/sru/47BIBSYS_NETWORK?version=1.2&operation=searchRetrieve&recordSchema=marcxml&query=alma.isbn="
			res, err := client.Open(ctx, url+id)
			if err != nil {
				return Ingestion{}, err
			}
			defer res.Close()
			var sruRes sruResponse
			if err := xml.NewDecoder(res).Decode(&sruRes); err != nil {
				return Ingestion{}, err
			} else if sruRes.NumberOfRecords == 0 {
				return Ingestion{}, sirkulator.ErrNotFound
			} else if sruRes.NumberOfRecords > 1 {
				// Bail out if we get more than one record
				return Ingestion{}, sirkulator.ErrConflict // TODO or craft custom error here
			}
			var mrc marc.Record
			if err := marc.Unmarshal(sruRes.Records.Record.Metadata, &mrc); err != nil {
				return Ingestion{}, err
			}
			return ingestMarcRecord("bibsys", mrc, idFunc)
		},
	},
}

// remoteRecord will go through the list of externalSources and try to get an
// Ingestion from the remote record. It will at most use one external source.
func (ig *Ingestor) remoteRecord(ctx context.Context, idtype, id string) (Ingestion, error) {
	for _, src := range externalSources {
		if data, err := src.Fetch(ctx, idtype, id, ig.idFunc); err == nil {
			// We return as soon as we have a valid response.
			// TODO (future idea) consider combining severeal remote records.
			return data, nil
		}
		// TODO which errors are interesting to callers? ErrTemporary - to signal it might
		// be worthwile to try again?
	}
	return Ingestion{}, sirkulator.ErrNotFound
}

// persistIngestion will store all resources, relations and reviews in
// the Ingestion. No further validation of input is performed - all of
// the given data is assumed to be valid at this point, as not to
// trigger any SQL constraint errors when inserting into DB.
//
// CreatedAt/UpdatdAt timestamps on resources will be set
// TODO consider setting them earlier
func persistIngestion(conn *sqlite.Conn, data Ingestion) (err error) {
	defer sqlitex.Save(conn)(&err)

	now := time.Now()

	for i, res := range data.Resources {
		stmt := conn.Prep(`
			INSERT INTO resource (type, id, label, data, created_at, updated_at)
				VALUES ($type, $id, $label, $data, $now, $now)
		`)

		stmt.SetText("$type", res.Type.String())
		stmt.SetText("$id", res.ID)
		stmt.SetText("$label", res.Label)
		stmt.SetInt64("$now", now.Unix())
		// TODO set created_at && update_at to time.Now.Unix()
		b, err := json.Marshal(res.Data)
		if err != nil {
			return err // TODO annotate
		}
		stmt.SetBytes("$data", b)
		if _, err := stmt.Step(); err != nil {
			return err // TODO annotate
		}

		// Update timestamps
		data.Resources[i].CreatedAt = now
		data.Resources[i].UpdatedAt = now

		// Persist resource links. Duplicate entries will be ignored.
		for _, link := range res.Links {
			stmt := conn.Prep(`
				INSERT OR IGNORE INTO link (resource_id, type, id)
					VALUES ($resource_id, $type, $id)
			`)
			stmt.SetText("$resource_id", res.ID)
			stmt.SetText("$type", link[0])
			stmt.SetText("$id", link[1])
			if _, err := stmt.Step(); err != nil {
				return err // TODO annotate
			}
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

// downloadImages will try to download an image and store it in DB, stopping
// at the first successfull operation.
// No errors are reported or logged.
func (ig *Ingestor) downloadImages(ctx context.Context, files []FileFetch) {
	conn := ig.db.Get(ctx)
	if conn == nil {
		return // context.Cancelled
	}
	defer ig.db.Put(conn)
	for _, f := range files {
		r, err := client.Open(ctx, f.URL)
		if err != nil {
			continue
		}
		src, _, err := image.Decode(r)
		if err != nil {
			r.Close()
			continue
		}
		r.Close()
		ratio := float64(ig.ImageWidth) / float64(src.Bounds().Max.X)
		height := float64(src.Bounds().Max.Y) * ratio
		dst := image.NewRGBA(image.Rect(0, 0, ig.ImageWidth, int(height)))
		draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
		var b bytes.Buffer
		if err := jpeg.Encode(&b, dst, nil); err != nil {
			continue
		}
		stmt := conn.Prep(`
			INSERT INTO files.image (id, type, width, height, size, data, source)
				VALUES ($id, $type, $width, $height, $size, $data, $source)`)
		stmt.SetText("$id", f.ResourceID)
		stmt.SetText("$type", "jpeg")
		stmt.SetInt64("$width", 200)
		stmt.SetInt64("$height", int64(height))
		stmt.SetInt64("$size", int64(b.Len()))
		stmt.SetBytes("$data", b.Bytes())
		stmt.SetText("$source", f.URL)
		if _, err = stmt.Step(); err != nil {
			continue
		}

		break // Sucessfully stored an image.
	}
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
		// We loop backwards, which make it easier to remove resource from data.Resources
		// if we find any matches.

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
			if err != nil && !errors.Is(err, sqlitex.ErrNoResults) {
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

	if ig.ImageDownload {
		if ig.ImageAsync {
			go ig.downloadImages(context.Background(), data.Covers)
		} else {
			ig.downloadImages(ctx, data.Covers)
		}
	}

	// Index documents asynchronously
	go ig.indexResources(data.Resources)

	return nil
}

func (ig *Ingestor) indexResources(res []sirkulator.Resource) {
	if ig.idx == nil {
		return
	}

	var docs []search.Document
	for _, r := range res {
		docs = append(docs, search.Document{
			ID:        r.ID,
			Type:      r.Type.String(),
			Label:     r.Label,
			Gain:      1.0,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}
	if err := ig.idx.Store(docs...); err != nil {
		log.Println(err) // TODO or not
	}
	// TODO update indexed_at column in main.resource SQL db
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
	URL        string
}
