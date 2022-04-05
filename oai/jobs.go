package oai

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/client"
)

type HarvestJob struct {
	Harvester
	JobName string
}

func (h *HarvestJob) Name() string {
	return h.JobName
}

func (h *HarvestJob) Run(ctx context.Context, w io.Writer) error {
	return h.Harvester.Run(ctx, w)
}

// Publisher is a entry in NB isbn publisher database (https://nb.no/isbnforlag)
// Not an oai source, but we store it as oai records for now.
// TODO consider expanding oai db to handle all externally data sources which
// are kept in sync locally
type Publisher struct {
	ID           string
	Name         string
	FullName     string
	AltName      string
	WWW          string
	ISBNPrefixes []string
	Notes        []string
	CreatedAt    string
	UpdatedAt    string
}

func (p Publisher) Resource() sirkulator.Resource {
	res := sirkulator.Resource{
		Type:      sirkulator.TypePublisher,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	pub := sirkulator.Publisher{
		Name:  p.Name,
		Notes: p.Notes,
	}
	if p.AltName != "" {
		pub.NameVariations = append(pub.NameVariations, p.AltName)
	}
	if p.FullName != "" {
		pub.NameVariations = append(pub.NameVariations, p.FullName)
	}
	if p.WWW != "" {
		res.Links = append(res.Links, [2]string{"www", p.WWW})
	}
	res.Links = append(res.Links, [2]string{"nb/isbnforlag", p.ID})
	for _, prefix := range p.ISBNPrefixes {
		parts := strings.Split(prefix, "-")
		if len(parts) == 3 {
			res.Links = append(res.Links, [2]string{"isbn/publisher", parts[2]})
		}
	}
	res.Data = pub
	res.Label = pub.Label()
	return res
}

type HarvestPublishersJob struct {
	DB *sqlitex.Pool
}

func (h *HarvestPublishersJob) Name() string {
	return "oai_harvest_isbnforlag"
}

func gzipPublsher(p Publisher) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	if err := json.NewEncoder(gz).Encode(p); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("gzipPublsher: %w", err)
	}

	return b.Bytes(), nil
}

func (h *HarvestPublishersJob) insertSource(ctx context.Context) error {
	conn := h.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer h.DB.Put(conn)

	const q = "INSERT OR IGNORE INTO oai.source (id, url, dataset, prefix) VALUES ('nb/isbnforlag','','','')"

	return sqlitex.ExecTransient(conn, q, nil)
}

func (h *HarvestPublishersJob) persist(ctx context.Context, batch []Publisher) (err error) {
	conn := h.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer h.DB.Put(conn)
	defer sqlitex.Save(conn)(&err)

	stmt := conn.Prep(`
		INSERT INTO oai.record (source_id, id, data, created_at, updated_at)
		VALUES ($source, $id, $data, $created, $updated)
	ON CONFLICT DO UPDATE
		SET data=$data, updated_at=$updated`)

	const datelayout = "2006-01-02"
	var links [][2]string // [2]{record_id, id}
	for _, p := range batch {
		data, err := gzipPublsher(p)
		if err != nil {
			return err
		}
		if p.CreatedAt == "" {
			p.CreatedAt = p.UpdatedAt
		}
		created, _ := time.Parse(datelayout, p.CreatedAt)
		updated, _ := time.Parse(datelayout, p.UpdatedAt)

		stmt.SetText("$source", "nb/isbnforlag")
		stmt.SetText("$id", p.ID)
		stmt.SetBytes("$data", data)
		stmt.SetInt64("$created", created.Unix())
		stmt.SetInt64("$updated", updated.Unix())

		if _, err = stmt.Step(); err != nil {
			return fmt.Errorf("persist: %w", err)
		}
		stmt.Reset()

		for _, prefix := range p.ISBNPrefixes {
			parts := strings.Split(prefix, "-")
			if len(parts) == 3 {
				links = append(links, [2]string{p.ID, parts[2]})
			}

		}
	}

	stmt = conn.Prep(`
		INSERT OR IGNORE INTO oai.link (source_id, record_id, type, id)
			VALUES ('nb/isbnforlag', $record_id, 'isbn/publisher', $id)
	`)

	for _, link := range links {
		stmt.SetText("$record_id", link[0])
		stmt.SetText("$id", link[1])

		if _, err = stmt.Step(); err != nil {
			return fmt.Errorf("persist: %w", err)
		}
		stmt.Reset()
	}

	return nil
}

func (h *HarvestPublishersJob) Run(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, "Fetching isbnforlag.ndjson\n")

	const url = "https://gist.githubusercontent.com/boutros/0985e0dc230997451897cbef59edd947/raw"
	f, err := client.Open(ctx, url)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := h.insertSource(ctx); err != nil {
		return err
	}
	dec := json.NewDecoder(f)
	const batchSize = 100
	batch := make([]Publisher, 0, batchSize)
	fmt.Fprintf(w, "Persisting records using batch size=%d\n", batchSize)
	num := 0
	for {
		var p Publisher
		err := dec.Decode(&p)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		batch = append(batch, p)
		if len(batch) == batchSize {
			fmt.Fprint(w, ".")
			if err := h.persist(ctx, batch); err != nil {
				return err
			}
			num += len(batch)
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		if err := h.persist(ctx, batch); err != nil {
			return err
		}
		num += len(batch)
	}

	fmt.Fprintf(w, "\nDone inserted/updated %d records\n", num)

	return nil
}
