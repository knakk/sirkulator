package oai

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator/marc"
)

// Harvester represents a OAI harvester, capable of harvesting records from
// a remote repository and sync to a local DB.
type Harvester struct {
	DB       *sqlitex.Pool
	Endpoint string
	Source   string
	Set      string
	Prefix   string
	StartAt  time.Time
	Token    string
	Process  ProcessFunc
	Enqueue  bool // if true, set queued_at timestamp, and store data in new_data column // TODO implement!
}

func (h *Harvester) Name() string {
	return fmt.Sprintf("oai_harvester:%s:%s", h.Source, h.Set)
}

func (h *Harvester) Run(ctx context.Context, w io.Writer) error {
	if err := h.storeSource(ctx); err != nil {
		return fmt.Errorf("Harvester.Run: %w", err)
	}
	const batchSize int = 1000
	recordUpserts := make([]ProcessedRecord, 0, batchSize)
	recordArchived := make([]ProcessedRecord, 0)
	identifiers := make([][4]string, 0, batchSize)

	for {
		records, err := h.fetchRecords(ctx)
		if err != nil {
			return fmt.Errorf("Harvester.Run: %w", err)
		}

		for _, rec := range records {
			prec, err := h.Process(rec)
			if err != nil {
				w.Write([]byte(err.Error() + "\n"))
				// TODO consider continue harvesting, and just write err to w
				//return fmt.Errorf("Harvester.Run: %w", err)
			}
			prec.Source = h.Source
			fmt.Printf("%s %s %s %s\t%v\n", prec.Source, prec.ID, prec.Type, prec.Label, prec.Identifiers)
			if prec.ArchivedAt.IsZero() {
				recordUpserts = append(recordUpserts, prec)
				for _, id := range prec.Identifiers {
					identifiers = append(identifiers, [4]string{prec.Source, prec.ID, id[0], id[1]})
				}
			} else {
				recordArchived = append(recordArchived, prec)
			}
			if len(recordUpserts) == batchSize {
				if err := h.storeRecords(ctx, recordUpserts, recordArchived, identifiers); err != nil {
					return fmt.Errorf("Harvester.Run: %w", err)
				}
				recordUpserts = make([]ProcessedRecord, 0, batchSize)
				recordArchived = make([]ProcessedRecord, 0)
				identifiers = make([][4]string, 0, batchSize)
			}
		}

		if err := h.updateSource(ctx); err != nil {
			return fmt.Errorf("Harvester.Run: %w", err)
		}
		if h.Token == "" {
			// ResumptionToken is empty, which means we have harvested all records.
			break
		}
	}

	// Write any batched up records not yet persistent
	if len(recordUpserts) > 0 || len(recordArchived) > 0 {
		if err := h.storeRecords(ctx, recordUpserts, recordArchived, identifiers); err != nil {
			return fmt.Errorf("Harvester.Run: %w", err)
		}
	}
	return nil
}

// UpdateRecords fetches one or more records from remote repoistory and stores
// them in DB, either updating an exsisting record, or creating a new one.
func (h *Harvester) UpdateRecords(ctx context.Context, ids ...string) error {
	if err := h.storeSource(ctx); err != nil {
		return fmt.Errorf("UpdateRecords: %w", err)
	}
	recordUpserts := make([]ProcessedRecord, 0, len(ids))
	recordArchived := make([]ProcessedRecord, 0)
	identifiers := make([][4]string, 0, len(ids))

	for _, id := range ids {
		url := fmt.Sprintf("%s?verb=GetRecord&identifier=%s", h.Endpoint, id)
		if h.Prefix != "" {
			url += "&metadataPrefix=" + h.Prefix
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("UpdateRecords: NewRequest(%s): %w", url, err)
		}
		req.Header.Set("Accept", "text/xml")

		c := &http.Client{
			Timeout: 20 * time.Second,
		}
		resp, err := c.Do(req)
		if err != nil {
			return fmt.Errorf("UpdateRecords: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("UpdateRecords: got HTTP status: %v", resp.Status)
		}
		var oaiResponse getRecordResponse
		dec := xml.NewDecoder(resp.Body)
		if err := dec.Decode(&oaiResponse); err != nil {
			return fmt.Errorf("UpdateRecords: XML decode: %w", err)
		}

		// Return an error if there is any OAI error in the response.
		// Possible error codes for the GetRecord verb: TODO
		if errCode := oaiResponse.Error.Code; errCode != "" {
			return fmt.Errorf("UpdateRecords: OAI error: %s", errCode)
		}

		prec, err := h.Process(oaiResponse.GetRecord.Record)
		if err != nil {
			return fmt.Errorf("UpdateRecords: %w", err)
		}
		if prec.ArchivedAt.IsZero() {
			recordUpserts = append(recordUpserts, prec)
			for _, id := range prec.Identifiers {
				identifiers = append(identifiers, [4]string{prec.Source, prec.ID, id[0], id[1]})
			}
		} else {
			recordArchived = append(recordArchived, prec)
		}
	}
	if err := h.storeRecords(ctx, recordUpserts, recordArchived, identifiers); err != nil {
		return fmt.Errorf("UpdateRecords: %w", err)
	}

	return nil
}

func (h *Harvester) fetchRecords(ctx context.Context) ([]RemoteRecord, error) {
	//return testRecords() // TODO remove
	url := h.Endpoint + "?verb=ListRecords"
	if h.Token != "" {
		url += "&resumptionToken=" + h.Token
	} else {
		url += "&metadataPrefix=" + h.Prefix
		if h.Set != "" {
			url += "&set=" + h.Set
		}
		if !h.StartAt.IsZero() {
			url += "&from=" + h.StartAt.Format("2006-01-02")
		}
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetchRecords: NewRequest(%s): %w", url, err)
	}
	req.Header.Set("Accept", "text/xml")

	c := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetchRecords: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetchRecords: got HTTP status: %v", resp.Status)
	}
	var oaiResponse listRecordsResponse
	dec := xml.NewDecoder(resp.Body)
	if err := dec.Decode(&oaiResponse); err != nil {
		return nil, fmt.Errorf("fetchRecords: XML decode: %w", err)
	}

	// Return an error if there is any OAI error in the response.
	// Possible error codes for the ListRecords verb:
	//	badArgument, badResumptionToken, noRecordsMatch, noSetHierarchy
	if errCode := oaiResponse.Error.Code; errCode != "" {
		return nil, fmt.Errorf("fetchRecords: OAI error: %s", errCode)
	}

	// Always store the resumptionToken on succesfull fetch, even if empty string.
	h.Token = oaiResponse.ListRecords.ResumptionToken

	return oaiResponse.ListRecords.Records, nil
}

func (h *Harvester) storeSource(ctx context.Context) error {
	conn := h.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer h.DB.Put(conn)
	const q = "INSERT OR IGNORE INTO oai_source (id, url, dataset, prefix) VALUES (?, ?, ?, ?)"
	if err := sqlitex.Exec(conn, q, nil, h.Source, h.Endpoint, h.Set, h.Prefix); err != nil {
		return fmt.Errorf("storeSource: %w", err)
	}
	return nil
}

func (h *Harvester) updateSource(ctx context.Context) error {
	conn := h.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer h.DB.Put(conn)
	if h.Token == "" {
		// TODO we should use a timestamp from remote repository as in_sync_at
		const q = "UPDATE oai_source SET token=?, in_sync_at=? WHERE id=?"
		if err := sqlitex.Exec(conn, q, nil, h.Token, time.Now().Unix(), h.Source); err != nil {
			return fmt.Errorf("updateSource: %w", err)
		}
		return nil
	}
	const q = "UPDATE oai_source SET token=?, in_sync_at=NULL WHERE id=?"
	if err := sqlitex.Exec(conn, q, nil, h.Token, h.Source); err != nil {
		return fmt.Errorf("updateSource: %w", err)
	}

	return nil
}

func (h *Harvester) storeRecords(ctx context.Context, upserts, archived []ProcessedRecord, identifiers [][4]string) (err error) {
	conn := h.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer h.DB.Put(conn)
	defer sqlitex.Save(conn)(&err)

	stmt, err := conn.Prepare(`
		INSERT INTO oai_record (source_id, id, data, created_at, updated_at, queued_at)
			VALUES ($source, $id, $data, $created, $updated, $queued)
		ON CONFLICT(source_id, id) DO UPDATE
			SET new_data=$data, queued_at=$queued
	`)
	if err != nil {
		return fmt.Errorf("storeRecords: %w", err)
	}
	for _, r := range upserts {
		stmt.SetText("$source", r.Source)
		stmt.SetText("$id", r.ID)
		stmt.SetBytes("$data", r.Data)
		stmt.SetInt64("$created", r.CreatedAt.Unix())
		stmt.SetInt64("$updated", r.UpdatedAt.Unix())
		stmt.SetInt64("$queued", time.Now().Unix())

		if _, err = stmt.Step(); err != nil {
			return fmt.Errorf("storeRecords: %w", err)
		}
		stmt.Reset()
	}

	stmt, err = conn.Prepare(`
		UPDATE OR IGNORE oai_record SET archived_at=$archived, queued_at=$queued
		WHERE source_id=$source AND id=$id
	`)
	if err != nil {
		return fmt.Errorf("storeRecords: %w", err)
	}

	for _, r := range archived {
		stmt.SetText("$source", r.Source)
		stmt.SetText("$id", r.ID)
		stmt.SetInt64("$archived", r.ArchivedAt.Unix())
		stmt.SetInt64("$queued", time.Now().Unix())

		if _, err = stmt.Step(); err != nil {
			return fmt.Errorf("storeRecords: %w", err)
		}
		stmt.Reset()
	}

	stmt, err = conn.Prepare(`
		INSERT OR IGNORE INTO oai_record_id (source_id, record_id, type, id)
		VALUES ($source_id, $record_id, $type, $id)
	`)
	if err != nil {
		return fmt.Errorf("storeRecords: %w", err)
	}
	for _, id := range identifiers {
		stmt.SetText("$source_id", id[0])
		stmt.SetText("$record_id", id[1])
		stmt.SetText("$type", id[2])
		stmt.SetText("$id", id[3])

		if _, err = stmt.Step(); err != nil {
			return fmt.Errorf("storeRecords: %w", err)
		}
		stmt.Reset()
	}
	return nil
}

func ProcessBibsys(rec RemoteRecord) (ProcessedRecord, error) {
	res := ProcessedRecord{}
	res.UpdatedAt = rec.Header.Datestamp
	if rec.Header.Status == "deleted" {
		res.ArchivedAt = rec.Header.Datestamp

		// Deleted records have no Metadata content
		return res, nil
	}
	var mrc marc.Record
	if err := marc.Unmarshal(rec.Metadata, &mrc); err != nil {
		return res, fmt.Errorf("ProcessBibsys(id=%s): decode marc error: %w", res.ID, err)
	}

	IndexBibsysAuthority(&res, mrc)

	b, err := gzipData([]byte(rec.Metadata))
	if err != nil {
		return res, fmt.Errorf("ProcessBibsys(id=%s):  %w", res.ID, err)
	}
	res.Data = b

	return res, nil
}

func gunzipData(data []byte) ([]byte, error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, fmt.Errorf("gunzipData: %w", err)
	}

	var res bytes.Buffer
	_, err = res.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("gunzipData: %w", err)
	}

	return res.Bytes(), nil
}

func gzipData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("gzipData: %w", err)
	}

	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("gzipData: %w", err)
	}

	return b.Bytes(), nil
}
