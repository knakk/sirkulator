package search

import (
	"context"
	"io"
	"time"

	"crawshaw.io/sqlite/sqlitex"
)

// Document represents a indexable document, or a document retrieved
// from index in a search results.
type Document struct {
	// Mandatory fields
	ID    string
	Type  string
	Label string

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
