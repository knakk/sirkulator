package oai

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator/marc"
)

// TODO: DocumentIndexer

// TODO rename to IdentifierIndexer
type Indexer struct {
	DB          *sqlitex.Pool
	Source      string
	Process     ProcessFunc
	ResetQueued bool
}

func (idx *Indexer) Name() string {
	return fmt.Sprintf("oai_indexer:%s", idx.Source)
}

func (idx *Indexer) Run(ctx context.Context, w io.Writer) error {
	const batchSize int = 1000
	identifiers := make([][4]string, 0, batchSize)
	record_ids := make([]string, 0, batchSize)

	for {
		records, err := idx.loadRecords(ctx, batchSize)
		if err != nil {
			return fmt.Errorf("Indexer.Run: %w", err)
		}
		if len(records) == 0 {
			break
		}
		for _, rec := range records {
			record_ids = append(record_ids, rec.ID)
			var mrc marc.Record
			if err := marc.Unmarshal(rec.Data, &mrc); err != nil {
				return fmt.Errorf("Indexer.Run: %w", err) // TODO write to w and continue
			}
			var res ProcessedRecord           // TODO this is cumbersome, we dont need ProcessedRecord here
			IndexBibsysPublication(&res, mrc) // <- except for this fn
			for _, id := range res.Identifiers {
				identifiers = append(identifiers, [4]string{rec.Source, rec.ID, id[0], id[1]})
			}
		}
		if err := idx.storeIdentifiers(ctx, identifiers, record_ids); err != nil {
			return fmt.Errorf("Indexer.Run: %w", err)
		}

		identifiers = make([][4]string, 0, batchSize)
		record_ids = make([]string, 0, batchSize)

	}
	return nil
}

func (idx *Indexer) loadRecords(ctx context.Context, num int) ([]DBRecord, error) {
	conn := idx.DB.Get(ctx)
	if conn == nil {
		return nil, context.Canceled
	}
	defer idx.DB.Put(conn)

	res := make([]DBRecord, 0, num)
	fn := func(stmt *sqlite.Stmt) error {
		blob, err := conn.OpenBlob("main", "oai_record", "data", stmt.ColumnInt64(2), false)
		if err != nil {
			return err
		}
		defer blob.Close()
		gz, err := gzip.NewReader(blob)
		if err != nil {
			return err
		}
		var b bytes.Buffer
		_, err = b.ReadFrom(gz)
		if err != nil {
			return err
		}
		// TODO stream parse marc record, maybe store Marc record, not bytes in Record stuct.
		res = append(res, DBRecord{
			Source: stmt.ColumnText(0),
			ID:     stmt.ColumnText(1),
			Data:   b.Bytes(),
		})
		return nil
	}

	const q = `SELECT source_id, id, rowid
				FROM oai_record
				WHERE archived_at IS NULL AND queued_at IS NOT NULL LIMIT ?`
	if err := sqlitex.Exec(conn, q, fn, num); err != nil {
		return nil, fmt.Errorf("loadRecords: %w", err)
	}

	return res, nil
}

func (idx *Indexer) storeIdentifiers(ctx context.Context, identifiers [][4]string, ids []string) (err error) {
	conn := idx.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer idx.DB.Put(conn)
	defer sqlitex.Save(conn)(&err)

	stmt, err := conn.Prepare(`
		INSERT OR IGNORE INTO oai_record_id (source_id, record_id, type, id)
			VALUES ($source_id, $record_id, $type, $id)
	`)
	if err != nil {
		return fmt.Errorf("storeIdentifiers: %w", err)
	}

	stmt2, err := conn.Prepare(`UPDATE oai_record SET queued_at=NULL WHERE source_id=$source_id AND id=$id`)
	if err != nil {
		return fmt.Errorf("storeIdentifiers: %w", err)
	}

	for _, id := range identifiers {
		stmt.SetText("$source_id", id[0])
		stmt.SetText("$record_id", id[1])
		stmt.SetText("$type", id[2])
		stmt.SetText("$id", id[3])

		if _, err = stmt.Step(); err != nil {
			return fmt.Errorf("storeIdentifiers: %w", err)
		}
		stmt.Reset()
	}
	for _, id := range ids {
		stmt2.SetText("$source_id", idx.Source)
		stmt2.SetText("$id", id)
		if _, err = stmt2.Step(); err != nil {
			return fmt.Errorf("storeIdentifiers: %w", err)
		}
		stmt2.Reset()
	}

	return nil
}

func IndexBibsysPublication(res *ProcessedRecord, mrc marc.Record) {
	for _, isbn := range mrc.ValuesAt("020", "a") {
		if len(isbn) < 10 { // TODO proper validation?
			continue
		}
		res.Identifiers = append(res.Identifiers, [2]string{"isbn", strings.Replace(isbn, "-", "", -1)})
	}

	for _, issn := range mrc.ValuesAt("022", "a") {
		if len(issn) < 8 { // TODO proper validation?
			continue
		}
		// TODO make sure all have same format:\d{4}-\d{4} or strip dash
		res.Identifiers = append(res.Identifiers, [2]string{"issn", issn})
	}

	for _, ismn := range mrc.ValuesAt("024", "a") {
		if len(ismn) < 13 { // TODO proper validation?
			continue
		}
		res.Identifiers = append(res.Identifiers, [2]string{"ismn", strings.Replace(ismn, "-", "", -1)})
	}

	for _, ean := range mrc.ValuesAt("025", "a") {
		if len(ean) < 13 { // TODO proper validation?
			continue
		}
		res.Identifiers = append(res.Identifiers, [2]string{"ean", strings.Replace(ean, "-", "", -1)})
	}

	if author, ok := mrc.ValueAt("100", "a"); ok {
		res.Label += invertName(author) + ": "
	} // TODO 100$c

	if title, ok := mrc.ValueAt("245", "a"); ok {
		res.Label += strings.TrimSuffix(strings.TrimSpace(title), ":")
	}
	if subtitle, ok := mrc.ValueAt("245", "b"); ok {
		res.Label += ": " + strings.TrimSpace(subtitle)
	}
	if year, ok := mrc.ValueAt("260", "c"); ok {
		res.Label += " (" + cleanYear(year) + ")"
	}
	// TODO 028$a serial number/catalogue number for music records/sheet music
	// https://www.iasa-web.org/cataloguing-rules/sound-record-catalogue-numbers
	// namme: catnum? muscat? muscatnum?
}

func IndexBibsysAuthority(res *ProcessedRecord, mrc marc.Record) {
	if cfield, ok := mrc.ControlFieldAt("001"); ok {
		res.ID = cfield.Value
	} // TODO handle no ID!

	if cfield, ok := mrc.ControlFieldAt("008"); ok && len(cfield.Value) >= 6 {
		t, err := time.Parse("060102", cfield.Value[0:6]) // Date Entered On File
		if err == nil {
			res.CreatedAt = t
		}
	}
	if _, ok := mrc.DataFieldAt("245"); ok {
		res.Type = "publication"
	} else if d, ok := mrc.DataFieldAt("100"); ok {
		res.Type = "person"
		res.Label = d.ValueAt("a")
		res.Label = invertName(res.Label)
		if num := d.ValueAt("b"); num != "" {
			res.Label = fmt.Sprintf("%s %s", res.Label, num)
		}
		if exp := d.ValueAt("c"); exp != "" {
			res.Label = fmt.Sprintf("%s, %s", res.Label, exp)
		}
		if dates := d.ValueAt("d"); dates != "" {
			res.Label = fmt.Sprintf("%s (%s)", res.Label, dates)
		}
	} else if d, ok := mrc.DataFieldAt("110"); ok {
		res.Type = "corporation"
		res.Label = d.ValueAt("a")
		if sub := d.ValueAt("b"); sub != "" {
			res.Label = fmt.Sprintf("%s > %s", res.Label, sub)
		}
	} else if d, ok := mrc.DataFieldAt("111"); ok {
		res.Type = "event"
		res.Label = d.ValueAt("a")
		if sub := d.ValueAt("b"); sub != "" {
			res.Label = fmt.Sprintf("%s > %s", res.Label, sub)
		}
	} else if d, ok := mrc.DataFieldAt("130"); ok {
		res.Type = "uniformtitle"
		res.Label = d.ValueAt("a")
		if sub := d.ValueAt("b"); sub != "" {
			res.Label = fmt.Sprintf("%s > %s", res.Label, sub)
		}
	} else {
		res.Type = "unknown"
	}
	if res.Type == "publication" {
		IndexBibsysPublication(res, mrc)
		return
	}

	for _, d := range mrc.DataFieldsAt("024") {
		code := d.ValueAt("2")
		val := d.ValueAt("a")
		if val == "" {
			continue
		}
		switch strings.ToLower(code) {
		case "viaf":
			res.Identifiers = append(res.Identifiers, [2]string{"viaf", strings.TrimPrefix(val, "http://viaf.org/viaf/")})
		case "isni":
			res.Identifiers = append(res.Identifiers, [2]string{"isni", val})
		case "bibbi":
			res.Identifiers = append(res.Identifiers, [2]string{"bibbi", val})
		case "no-trbib", "dma", "hdl":
			// ignore
		default:
			fmt.Printf("unhandled 024 identifier: %s\t%s\n", code, val)
		}
	}
}

func cleanYear(s string) string {
	// TODO årsspenn for tidsskrifter, ex: "2002-[2009]""
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	s = strings.TrimPrefix(s, "c")
	return s
}

func invertName(s string) string {
	if i := strings.Index(s, ", "); i != -1 {
		return s[i+2:] + " " + s[:i]
	}
	return s
}
