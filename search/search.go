package search

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blugelabs/bluge"
)

// Document represents a indexable document, or a document retrieved
// from index in a search results.
type Document struct {
	// Mandatory fields
	ID    string
	Type  string
	Label string
	Gain  float64

	// Other fields ([0] = key, [1] = value)
	Fields [][2]string
}

func (d Document) GetField(key string) (string, bool) {
	for _, f := range d.Fields {
		if f[0] == key {
			return f[1], true
		}
	}
	return "", false
}

func (d Document) GetAllOfField(key string) (res []string) {
	for _, f := range d.Fields {
		if f[0] == key {
			res = append(res, f[1])
		}
	}
	return res
}

// Index represents a search index that can index and query Documents.
type Index struct {
	writer *bluge.Writer
}

// Open creates the Index on disk and make it ready for indexing and querying.
func Open(dir string) (*Index, error) {
	// Create directory if it doesn't exist.
	dir = strings.TrimSuffix(dir, "/")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("search: Open(%s): MkdirAll: %w", dir, err)
	}

	conf := bluge.DefaultConfig(dir)
	w, err := bluge.OpenWriter(conf)
	if err != nil {
		return nil, fmt.Errorf("search: Open(%s): %w", dir, err)
	}

	return &Index{writer: w}, nil
}

// OpenMem creates an in-memory index usefull for testing
func OpenMem() (*Index, error) {
	w, err := bluge.OpenWriter(bluge.InMemoryOnlyConfig())
	if err != nil {
		return nil, fmt.Errorf("search: OpenMem: %w", err)
	}

	return &Index{writer: w}, nil
}

func (idx *Index) Close() error {
	if idx.writer == nil {
		return nil
	}
	return idx.writer.Close() // TODO annotate
}

func (idx *Index) Store(docs ...Document) error {
	const batchThreshold = 5 // TODO measure
	if len(docs) > batchThreshold {
		return idx.batchStore(docs)
	}
	for _, doc := range docs {
		d := bluge.NewDocument(doc.ID).
			AddField(bluge.NewTextField("type", doc.Type).SearchTermPositions().StoreValue()).
			AddField(bluge.NewTextField("label", doc.Label).SearchTermPositions().StoreValue()).
			AddField(bluge.NewNumericField("gain", doc.Gain))
		if err := idx.writer.Update(d.ID(), d); err != nil {
			return fmt.Errorf("search: Index.Store: writing doc %s: %w", d.ID(), err)
		}
	}

	return nil
}

func (idx Index) batchStore(docs []Document) error {
	batch := bluge.NewBatch()

	// TODO verify docs does not contain duplicate IDs?
	for _, doc := range docs {
		d := bluge.NewDocument(doc.ID).
			AddField(bluge.NewTextField("type", doc.Type).SearchTermPositions().StoreValue()).
			AddField(bluge.NewTextField("label", doc.Label).SearchTermPositions().StoreValue()).
			AddField(bluge.NewNumericField("gain", doc.Gain))
		batch.Update(d.ID(), d)
	}

	if err := idx.writer.Batch(batch); err != nil {
		return fmt.Errorf("search: Index.batchStore: writing batch %w", err)
	}
	return nil
}

func (idx *Index) Search(ctx context.Context, q string, limit int) (Results, error) {
	res := Results{}
	var query bluge.Query
	if q == "" {
		query = bluge.NewMatchAllQuery()
	} else {
		query = bluge.NewFuzzyQuery(q).SetField("label")
	}
	req := bluge.NewTopNSearch(limit, query).WithStandardAggregations()
	r, _ := idx.writer.Reader() // err is always nil: https://github.com/blugelabs/bluge/issues/35
	defer r.Close()             // TODO catch and return error
	dmi, err := r.Search(ctx, req)
	if err != nil {
		return res, fmt.Errorf("search: Index.Search: search: %w", err)
	}
	res.Total = dmi.Aggregations().Count()
	res.Time = dmi.Aggregations().Duration()

	// Iterate through the query matches
	match, err := dmi.Next()
	for err == nil && match != nil {
		hit := Hit{Score: match.Score}
		err = match.VisitStoredFields(func(field string, value []byte) bool {
			// TODO or use match.DocValues?
			switch field {
			case "_id":
				hit.ID = string(value)
			case "type":
				hit.Type = string(value)
			case "label":
				hit.Label = string(value)
			}
			return true
		})
		if err != nil {
			return res, fmt.Errorf("search: Index.Search: loading fields: %w", err)
		}
		res.Hits = append(res.Hits, hit)

		match, err = dmi.Next() // load next match
	}
	if err != nil {
		return res, fmt.Errorf("search: Index.Search: iterating results: %w", err)
	}

	return res, nil
}

type Results struct {
	Hits  []Hit
	Total uint64
	Time  time.Duration
}

type Hit struct {
	Document
	Score float64
}

/*
func DefaultBatchChan() chan<- Document {
	return defaultBatchChan
}

var (
	defaultBatchSize = 100
	defaultBatchWait = time.Second * 2
	defaultBatchChan = make(chan Document, defaultBatchSize)
)

type Indexer struct {
}

func (i Indexer) Run(ctx context.Context, db *sqlitex.Pool, w io.Writer) {
	for {

	}
}

func Index(ctx context.Context, db *sqlitex.Pool, docs []Document) error {
	return nil
}

*/
