package search

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
)

type Indexer struct {
	wg sync.WaitGroup

	DB        *sqlitex.Pool
	Idx       *Index
	BatchSize int
}

func (i *Indexer) Name() string {
	return "reindex_all_resources"
}

func (i *Indexer) Run(ctx context.Context, w io.Writer) error {
	conn := i.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer i.DB.Put(conn)

	var rowid int64
	hasMore := true
	q := `
		SELECT
			rowid,
			id,
			type,
			label,
			gain,
			created_at,
			updated_at,
			archived_at
		FROM resource
		WHERE rowid > ?
		ORDER BY rowid ASC
		LIMIT ?`

	docs := make([]Document, 0, i.BatchSize)
	stats := make(map[string]int)

	fn := func(stmt *sqlite.Stmt) error {
		hasMore = true
		rowid = stmt.ColumnInt64(0)

		var doc Document
		doc.ID = stmt.ColumnText(1)
		doc.Type = stmt.ColumnText(2)
		doc.Label = stmt.ColumnText(3)
		doc.Gain = stmt.ColumnFloat(4)
		doc.CreatedAt = time.Unix(stmt.ColumnInt64(5), 0)
		doc.CreatedAt = time.Unix(stmt.ColumnInt64(6), 0)
		if archived := stmt.ColumnInt64(7); archived != 0 {
			doc.ArchivedAt = time.Unix(archived, 0)
		}

		docs = append(docs, doc)
		stats[doc.Type]++

		return nil
	}

	fmt.Fprintf(w, "Starting indexing with batchsize=%d\n", i.BatchSize)

	for hasMore {
		hasMore = false
		if err := sqlitex.Exec(conn, q, fn, rowid, i.BatchSize); err != nil {
			return err
		}

		if len(docs) > 0 {
			fmt.Fprint(w, ".")
			i.wg.Add(1)
			d := make([]Document, 0, len(docs))
			copy(d, docs)
			go func() {
				i.Idx.batchStore(d)
				i.wg.Done()
			}()
			docs = docs[:0]
		}
	}

	i.wg.Wait()

	total := 0

	fmt.Fprintln(w, "\nDone indexing.\n\nStats by resource type:")

	for k, n := range stats {
		fmt.Fprintf(w, "%d\t%s\n", n, k)
		total += n
	}

	fmt.Fprintf(w, "\n%d\tsum total resources", total)

	return nil
}
