// package etl
package etl

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"image"
	"image/jpeg"
	"log"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/client"
	"github.com/knakk/sirkulator/isbn"
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

type Ingestion struct {
	Resources []sirkulator.Resource
	Relations []sirkulator.Relation
	Covers    []FileFetch
}

type FileFetch struct {
	ResourceID string
	URL        string
}

func NewIngestor(db *sqlitex.Pool, idx *search.Index) *Ingestor {
	return &Ingestor{
		db:         db,
		idx:        idx,
		ImageWidth: 300,                 // default resize width
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

type ImportEntry struct {
	Source    string // bibsys, openlibrary, discogs etc
	Exists    bool   // allready in local DB
	Error     string // not found, unable to map to local metadata, insuficcient data etc
	Resources []sirkulator.SimpleResource
}

// IngestISBN will try to ingest a publication and related resources given an ISBN-number,
// optionally persisting it to DB.
// TODO refactor to remove duplication
func (ig *Ingestor) IngestISBN(ctx context.Context, id string, persist bool) ImportEntry {
	id = isbn.Clean(id)

	var entry ImportEntry

	// 1. First check if we allready have it:
	if res, err := ig.existingPublication(ctx, "isbn", id); err == nil {
		// ISBN is present on existing publication, so we can return with
		// reference to that
		entry.Source = "local"
		entry.Exists = true
		entry.Resources = append(entry.Resources, res)
		return entry
	} else if !errors.Is(err, sirkulator.ErrNotFound) {
		// i/o error or other internal problem
		log.Printf("Ingestor.IngestISBN: %v", err)
		entry.Error = sirkulator.ErrInternal.Code
		return entry
	}

	// 2. Next, look for it in local oai DB:
	rec, err := ig.localRecord(ctx, "isbn", id)
	if err == nil {
		entry.Source = rec.Source
		data, err := ingestMarcRecord(rec.Source, rec.Data, ig.idFunc)
		if errors.Is(err, sirkulator.ErrNotFound) {
			entry.Error = err.Error()
			return entry
		} else if err != nil {
			log.Printf("Ingestor.IngestISBN: %v", err)
			entry.Error = sirkulator.ErrInternal.Code
			return entry
		}
		res, err := ig.Ingest(ctx, data, persist)
		if err != nil {
			log.Printf("Ingestor.IngestISBN: %v", err)
			entry.Error = sirkulator.ErrInternal.Code
			return entry
		}
		entry.Resources = res
		return entry
	} else if !errors.Is(err, sirkulator.ErrNotFound) {
		// i/o error or other internal problem
		log.Printf("Ingestor.IngestISBN: %v", err)
		entry.Error = sirkulator.ErrInternal.Code
		return entry
	}

	// 3. Finally, search external sources:
	data, err := ig.remoteRecord(ctx, "isbn", id)
	if errors.Is(err, sirkulator.ErrNotFound) {
		entry.Error = sirkulator.ErrNotFound.Code
		return entry
	} else if err != nil {
		log.Printf("Ingestor.IngestISBN: %v", err)
		entry.Error = sirkulator.ErrInternal.Code
		return entry
	}
	res, err := ig.Ingest(ctx, data, persist)
	if err != nil {
		log.Printf("Ingestor.IngestISBN: %v", err)
		entry.Error = sirkulator.ErrInternal.Code
		return entry
	}
	entry.Resources = res

	return entry
}

func (ig *Ingestor) existingPublication(ctx context.Context, idtype, id string) (sirkulator.SimpleResource, error) {
	var res sirkulator.SimpleResource
	conn := ig.db.Get(ctx)
	if conn == nil {
		return res, context.Canceled
	}
	defer ig.db.Put(conn)

	fn := func(stmt *sqlite.Stmt) error {
		res.ID = stmt.ColumnText(0)
		res.Label = stmt.ColumnText(1)
		return nil
	}
	q := `
		SELECT
			r.id,
			r.label
		FROM resource r
			JOIN link l ON (l.resource_id=r.id AND l.type=? AND l.id=?)
	`
	if err := sqlitex.Exec(conn, q, fn, idtype, id); err != nil {
		return res, err
	}
	if res.ID == "" {
		return res, sirkulator.ErrNotFound
	}

	res.Type = sirkulator.TypePublication
	return res, nil
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
		FROM oai.link t
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

// persistIngestion will store all resources and relations in
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
			if link[0] == "isbn" {
				// TODO move out this to a fn enforcing correct format of all IDs
				stmt.SetText("$id", isbn.Clean(link[1]))
			} else {
				stmt.SetText("$id", link[1])
			}
			if _, err := stmt.Step(); err != nil {
				return err // TODO annotate
			}
		}
	}

	for _, rel := range data.Relations {
		stmt := conn.Prep(`
			WITH v(from_id, type, data) AS (VALUES ($from_id, $type, $data))
			INSERT INTO relation (from_id, to_id, type, data, queued_at)
			SELECT
				v.from_id,
				res.id,
				v.type,
				JSON_PATCH(v.data, IIF(res.id IS NULL,
					IIF($to_id != '', JSON_OBJECT('label', $to_id), '{}'), '{}')) AS data,
				IIF(res.id IS NULL, $queued_at, NULL) AS queued_at
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

// downloadImages will try to download an image and store it in DB, stopping
// at the first successfull operation.
// No errors are reported or logged.
func (ig *Ingestor) downloadImages(ctx context.Context, files []FileFetch) {
	conn := ig.db.Get(ctx)
	if conn == nil {
		return // context.Cancelled
	}
	defer ig.db.Put(conn)
	for _, f := range files { // TODO consider extract loop body to function, for easier cleanup with defer
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

		// TODO enforce minimum image size/quality?
		// const minWidth = 300
		// if src.Bounds().Max.X < minWidth {
		// 	continue
		// }

		ratio := float64(ig.ImageWidth) / float64(src.Bounds().Max.X)
		height := float64(src.Bounds().Max.Y) * ratio
		dst := image.NewRGBA(image.Rect(0, 0, ig.ImageWidth, int(height)))
		draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

		var b bytes.Buffer
		opt := jpeg.Options{Quality: 50} // TODO find optimal Quality

		if err := jpeg.Encode(&b, dst, &opt); err != nil {
			continue
		}

		stmt := conn.Prep(`
			INSERT INTO files.image (id, type, width, height, size, data, source)
				VALUES ($id, $type, $width, $height, $size, $data, $source)`)
		stmt.SetText("$id", f.ResourceID)
		stmt.SetText("$type", "jpeg")
		stmt.SetInt64("$width", int64(ig.ImageWidth))
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

// Ingest will merge the ingestion with locally matching resources before storing the data to db
// and trigger indexing of documents.
// if persist=false, nothing is persisted, and the returnet results represents a preview of which resources
// would have been stored if persist=true.
func (ig *Ingestor) Ingest(ctx context.Context, data Ingestion, persist bool) ([]sirkulator.SimpleResource, error) {
	conn := ig.db.Get(ctx)
	if conn == nil {
		return nil, context.Canceled
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
		newResource := true
		stmt := conn.Prep("SELECT resource_id FROM link WHERE type=$type AND id=$id")
		for _, link := range res.Links {
			stmt.SetText("$type", link[0])
			stmt.SetText("$id", link[1])
			id, err := sqlitex.ResultText(stmt)
			if err != nil && !errors.Is(err, sqlitex.ErrNoResults) {
				return nil, err // TODO annotate
			}

			if id == "" {
				continue
			}
			// Resource is already in our DB and sholdn't be imported
			newResource = false
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

		if newResource {
			// check if an authority record is present in local oai db, and use that instead
			for _, link := range res.Links {
				if link[0] != "bibsys" {
					continue
				}
				var rec oai.Record
				fn := func(stmt *sqlite.Stmt) error {
					rec.Source = "bibsys/aut"
					rec.ID = link[1]
					blob, err := conn.OpenBlob("oai", "record", "data", stmt.ColumnInt64(0), false)
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
				q := "SELECT rowid FROM oai.record WHERE source_id='bibsys/aut' AND id=?"
				if err := sqlitex.Exec(conn, q, fn, link[1]); err != nil {
					return nil, err // TODO annotate
				}
				if rec.ID == "" {
					continue
				}
				switch res.Type {
				case sirkulator.TypePerson:
					p, err := PersonFromAuthority(rec.Data)
					if err == nil {
						// Swap resource with the oai resource,
						// making sure to copy the ID from the discarded resource.
						data.Resources[i] = p
						data.Resources[i].ID = res.ID
					} else {
						log.Printf("Ingesor.Ingest: %v", err)
					}
				case sirkulator.TypeCorporation:
					c, err := CorporationFromAuthority(rec.Data)
					if err == nil {
						// Swap resource with the oai resource,
						// making sure to copy the ID from the discarded resource.
						data.Resources[i] = c
						data.Resources[i].ID = res.ID
					} else {
						log.Printf("Ingesor.Ingest: %v", err)
					}
					if parent := c.Data.(sirkulator.Corporation).ParentName; parent != "" {
						// Add review to establish link to parent corporation
						data.Relations = append(data.Relations, sirkulator.Relation{
							FromID: res.ID,
							Type:   "has_parent",
							Data:   map[string]any{"name": parent},
						})
					}
				}
			}
		}
	}

	results := make([]sirkulator.SimpleResource, 0, len(data.Resources))
	for _, r := range data.Resources {
		results = append(results, sirkulator.SimpleResource{
			Type:  r.Type,
			ID:    r.ID,
			Label: r.Label,
		})
	}

	// We're only interested in a preview, so return now before persiting anything.
	if !persist {
		return results, nil
	}

	// Store all resources and relations in a transaction:
	if err := persistIngestion(conn, data); err != nil {
		return nil, err // TODO annotate
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

	return results, nil
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
}
