package sql

import (
	"context"
	"fmt"
	"io"
	"time"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator/search"
)

// The  queries must follow the following formula:
// [0] text description of query
// [1] CTE named candidates
// [2] UPDATE query using candidates CTE

var janitorQueries = [][3]string{
	{
		"add missing to_id for relation of type 'has_classification' where label matches resource.id of type 'dewey'",
		`
			WITH candidates AS (
				SELECT rel.id, json_extract(rel.data, '$.label') AS dewey
				FROM relation rel JOIN resource res ON (json_extract(rel.data, '$.label') = res.id AND res.type='dewey')
				WHERE rel.type='has_classification' AND rel.to_id IS NULL
			)
		`,
		`
			UPDATE relation
			   SET to_id=json_extract(data, '$.label'),
			       queued_at = NULL
		      WHERE id IN (SELECT id FROM candidates)
		`,
	},
}

type JanitorJob struct {
	DB  *sqlitex.Pool
	Idx *search.Index
}

func (j *JanitorJob) Name() string {
	return "db_janitor"
}

func (j *JanitorJob) Run(ctx context.Context, w io.Writer) error {
	conn := j.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer j.DB.Put(conn)

	const qCount = "SELECT count(*) AS n FROM candidates"

	for i, q := range janitorQueries {
		fmt.Fprintf(w, "%d\t%s\n", i+1, q[0])

		stmt, _, err := conn.PrepareTransient(q[1] + qCount)
		if err != nil {
			return err // TODO annotate
		}
		if ok, err := stmt.Step(); err != nil {
			return err // TODO annotate
		} else if !ok {
			return fmt.Errorf("no rows returned for query %d", i+1)
		}

		n := stmt.GetInt64("n")
		fmt.Fprintf(w, "\tnumber of candidates: %d\n", n)

		stmt.Finalize()

		if n == 0 {
			// no candidates, so no point in going further with update query
			continue
		}

		t := time.Now()
		if err = sqlitex.Exec(conn, q[1]+q[2], nil); err != nil {
			return err // TODO annotate
		}
		dur := time.Since(t)
		fmt.Fprintf(w, "\tOK, update done in %v\n", dur)

	}

	return nil
}
