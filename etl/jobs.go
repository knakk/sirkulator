package etl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator/marc"
	"github.com/knakk/sparql"
)

type HarvestNBLinksJob struct {
	DB *sqlitex.Pool
}

func (h *HarvestNBLinksJob) Name() string {
	return "harvest_nb_links"
}

type nbResults struct {
	Page struct {
		TotalElements int `json:"totalElements"`
	} `json:"page"`
	Embedded struct {
		Items []struct {
			ID         string `json:"id"`
			AccessInfo struct {
				AccessAllowedFrom string `json:"accessAllowedFrom"` // NORWAY | NB | EVERYWHERE
			} `json:"accessInfo"`
			Metadata struct {
				Identifiers struct {
					OaiID string `json:"oaiId"`
					Urn   string `json:"urn"`
				} `json:"identifiers"`
			} `json:"metadata"`
		} `json:"items"`
	} `json:"_embedded"`
}

type bibsysResults []struct {
	XMLPresentation string `json:"xmlPresentation"`
}

func nbSearchViaBibsys(id string) (nbResults, error) {
	var res nbResults

	url := "https://api.bibs.aws.unit.no/alma?mms_id=" + id
	bsResp, err := http.Get(url)
	if err != nil {
		return res, nil
	}

	var bsRes bibsysResults
	if err := json.NewDecoder(bsResp.Body).Decode(&bsRes); err != nil {
		return res, err
	}
	if len(bsRes) != 1 {
		return res, nil
	}
	dec := marc.NewDecoder(bytes.NewBufferString(bsRes[0].XMLPresentation))
	rec, err := dec.Decode()
	if err != nil {
		return res, err
	}

	for _, f := range rec.DataFieldsAt("852") {
		/*
			<datafield ind1="0" ind2="1" tag="852">
				<subfield code="a">47BIBSYS_NB</subfield>
				<subfield code="6">991443004004702202</subfield>
				<subfield code="9">D</subfield>
			</datafield>
		*/
		if f.ValueAt("a") == "47BIBSYS_NB" {
			oaiID := f.ValueAt("6")
			if !strings.HasPrefix(oaiID, "99") {
				return res, nil
			}
			url := "https://api.nb.no/catalog/v1/items?q=oaiid:%22oai:nb.bibsys.no:" + oaiID + "%22"
			resp, err := http.Get(url)
			if err != nil {
				return res, err
			}
			defer resp.Body.Close()

			if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
				return res, err
			}
		}
	}

	return res, nil
}

func nbSearch(conn *sqlite.Conn, resourceID, bibsysID string, c, n *int) error {
	*c++
	url := "https://api.nb.no/catalog/v1/items?q=oaiid:%22oai:nb.bibsys.no:" + bibsysID[:len(bibsysID)-1] + "2%22"

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var res nbResults
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}

	if res.Page.TotalElements != 1 {
		res, err = nbSearchViaBibsys(bibsysID)
		if err != nil {
			return err
		}
		if res.Page.TotalElements != 1 {
			return nil
		}
	}

	var linkType string
	switch res.Embedded.Items[0].AccessInfo.AccessAllowedFrom {
	case "NORWAY":
		linkType = "nb/norway"
	case "EVERYWHERE":
		linkType = "nb/free"
	case "NB":
		linkType = "nb/restricted"
	default:
		return nil
	}

	if res.Embedded.Items[0].Metadata.Identifiers.Urn == "" {
		return nil
	}

	const q = `INSERT OR IGNORE INTO link (resource_id, type, id) VALUES (?, ?, ?)`

	if err := sqlitex.Exec(conn, q, nil, resourceID, linkType, res.Embedded.Items[0].Metadata.Identifiers.Urn); err != nil {
		return err
	}
	*n++
	return nil
}

func (h *HarvestNBLinksJob) Run(ctx context.Context, w io.Writer) error {
	conn := h.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer h.DB.Put(conn)

	// TODO filter out non-norwegian candidates, which are not on nb anyway
	const q = `
		SELECT l.resource_id, l.id FROM link l
			LEFT JOIN link l2 ON (l2.resource_id = l.resource_id AND l2.type IN ('nb/free','nb/norway','nb/restricted'))
		WHERE l.type='bibsys/pub' AND l2.id IS NULL
	`

	c := 0
	n := 0
	fn := func(stmt *sqlite.Stmt) error {
		return nbSearch(conn, stmt.ColumnText(0), stmt.ColumnText(1), &c, &n)
	}

	if err := sqlitex.Exec(conn, q, fn); err != nil {
		return err
	}

	fmt.Fprintf(w, "Done. Of %d candidates without link to nb.no added %d links.\n", c, n)

	return nil
}

type HarvestSNLLinksJob struct {
	DB *sqlitex.Pool
}

func (h *HarvestSNLLinksJob) Name() string {
	return "harvest_snl_links"
}

func (h *HarvestSNLLinksJob) Run(ctx context.Context, w io.Writer) error {
	conn := h.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer h.DB.Put(conn)

	const q = `
		SELECT l.resource_id, l.id FROM link l
			LEFT JOIN link l2 ON (l2.resource_id = l.resource_id AND l2.type = 'snl')
		WHERE l.type='bibsys/aut' AND l2.id IS NULL
	`

	wikidata, err := sparql.NewRepo("https://query.wikidata.org/sparql")
	if err != nil {
		return err
	}

	c := 0
	n := 0
	fn := func(stmt *sqlite.Stmt) error {
		return snlSearch(conn, wikidata, stmt.ColumnText(0), stmt.ColumnText(1), &c, &n)
	}

	if err := sqlitex.Exec(conn, q, fn); err != nil {
		return err
	}

	fmt.Fprintf(w, "Done. Of %d candidates without link to snl.no added %d links.\n", c, n)

	return nil
}

func snlSearch(conn *sqlite.Conn, repo *sparql.Repo, resourceID, bibsysID string, c, n *int) error {
	*c++

	const sparqlQ = `
		PREFIX wdt: <http://www.wikidata.org/prop/direct/>
		PREFIX wdtn: <http://www.wikidata.org/prop/direct-normalized/>

		# wdtn:P1015	NORAF-ID (tidligere bibsys-id)
		# wdt:P4342		Store norske leksikon-ID

		SELECT ?snl_id WHERE {
			?r 	wdtn:P1015 <https://livedata.bibsys.no/authority/%s> ;
				wdt:P4342 ?snl_id
		}
	`
	res, err := repo.Query(fmt.Sprintf(sparqlQ, bibsysID))
	if err != nil {
		return err
	}

	if len(res.Bindings()["snl_id"]) != 1 {
		return nil
	}

	snlID := res.Bindings()["snl_id"][0].String()

	const q = `INSERT OR IGNORE INTO link (resource_id, type, id) VALUES (?, ?, ?)`

	if err := sqlitex.Exec(conn, q, nil, resourceID, "snl", snlID); err != nil {
		return err
	}

	*n++

	return nil
}

type HarvestSNLDescriptions struct {
	DB *sqlitex.Pool
}

func (h *HarvestSNLDescriptions) Name() string {
	return "harvest_snl_descriptions"
}

func (h *HarvestSNLDescriptions) Run(ctx context.Context, w io.Writer) error {
	conn := h.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer h.DB.Put(conn)

	const q = `
		SELECT l.resource_id, l.id FROM link l
			LEFT JOIN resource_text rt ON (rt.resource_id = l.resource_id AND rt.source = 'snl')
		WHERE l.type='snl' AND rt.id IS NULL
	`

	c := 0
	n := 0
	fn := func(stmt *sqlite.Stmt) error {
		return snlDescription(conn, stmt.ColumnText(0), stmt.ColumnText(1), &c, &n)
	}

	if err := sqlitex.Exec(conn, q, fn); err != nil {
		return err
	}

	fmt.Fprintf(w, "Done. Of %d candidates without descriptions from snl.no added %d\n", c, n)

	return nil
}

type snlSearchResult struct {
	ID          string `json:"permalink"`
	Description string `json:"first_two_sentences"`
}

const snlAPIURL = "https://snl.no/api/v1/search?query="

func snlDescription(conn *sqlite.Conn, resourceID, snlID string, c, n *int) error {
	*c++
	snlIDEscaped := url.QueryEscape(snlID)
	req, err := http.NewRequest(http.MethodGet, snlAPIURL+snlIDEscaped, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err // TODO: annotate
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("querying snl.no/api/v1/search: http status %d", resp.StatusCode)
	}

	var snlRes []snlSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&snlRes); err != nil {
		return err // TODO: annotate
	}
	for _, r := range snlRes {
		if r.ID == snlID || r.ID == snlIDEscaped {

			const q = `
				INSERT INTO resource_text (resource_id, text, source, source_url, updated_at)
				VALUES (?, ?, 'snl', ?, ?)
			`

			if err := sqlitex.Exec(conn, q, nil, resourceID, r.Description, "https://snl.no/"+snlID, time.Now().Unix()); err != nil {
				return err
			}
			*n++
		}
	}

	return nil
}

type UpdateSNLDescriptions struct {
	DB *sqlitex.Pool
}

func (u *UpdateSNLDescriptions) Name() string {
	return "update_snl_descriptions"
}

func (u *UpdateSNLDescriptions) Run(ctx context.Context, w io.Writer) error {
	conn := u.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer u.DB.Put(conn)

	const q = `
		SELECT l.id, rt.id, rt.text FROM link l
			LEFT JOIN resource_text rt ON (rt.resource_id = l.resource_id AND rt.source = 'snl')
		WHERE l.type='snl'
	`

	c := 0
	n := 0
	fn := func(stmt *sqlite.Stmt) error {
		return snlUpdateDescription(conn, stmt.ColumnText(0), stmt.ColumnInt64(0), stmt.ColumnText(1), &c, &n)
	}

	if err := sqlitex.Exec(conn, q, fn); err != nil {
		return err
	}

	fmt.Fprintf(w, "Done. Of %d snl descriptions updated %d\n", c, n)

	return nil
}

func snlUpdateDescription(conn *sqlite.Conn, snlID string, textID int64, text string, c, n *int) error {
	*c++
	snlIDEscaped := url.QueryEscape(snlID)
	req, err := http.NewRequest(http.MethodGet, snlAPIURL+snlIDEscaped, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err // TODO: annotate
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("querying snl.no/api/v1/search: http status %d", resp.StatusCode)
	}

	var snlRes []snlSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&snlRes); err != nil {
		return err // TODO: annotate
	}
	for _, r := range snlRes {
		if r.ID == snlID || r.ID == snlIDEscaped {

			stmt := conn.Prep(`
				UPDATE resource_text
					SET text=$text, updated_at=$updated
				WHERE id=$id AND text != $text
				RETURNING id
			`)
			stmt.SetText("$text", r.Description)
			stmt.SetInt64("$updated", time.Now().Unix())
			stmt.SetInt64("$id", textID)
			if ok, err := stmt.Step(); err != nil {
				return err // TODO annotate
			} else if ok {
				*n++
			}
			stmt.Reset()
		}
	}

	return nil
}

type HarvestWikipediaLinks struct {
	DB *sqlitex.Pool
}

func (h *HarvestWikipediaLinks) Name() string {
	return "harvest_wikipedia_links"
}

func (h *HarvestWikipediaLinks) Run(ctx context.Context, w io.Writer) error {
	conn := h.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer h.DB.Put(conn)

	const q = `
		SELECT l.resource_id, l.id FROM link l
			LEFT JOIN link l2 ON (l2.resource_id = l.resource_id AND l2.type = 'wikidata')
		WHERE l.type='bibsys/aut' AND l2.id IS NULL
	`

	wikidata, err := sparql.NewRepo("https://query.wikidata.org/sparql")
	if err != nil {
		return err
	}

	c := 0
	n := 0
	fn := func(stmt *sqlite.Stmt) error {
		return wikipediaSearch(conn, wikidata, stmt.ColumnText(0), stmt.ColumnText(1), &c, &n)
	}

	if err := sqlitex.Exec(conn, q, fn); err != nil {
		return err
	}

	fmt.Fprintf(w, "Done. Of %d candidates without link to wikidata/wikipedia added %d links.\n", c, n)

	return nil
}

func wikipediaSearch(conn *sqlite.Conn, repo *sparql.Repo, resourceID, bibsysID string, c, n *int) error {
	*c++

	const sparqlQ = `
		PREFIX schema: <http://schema.org/>
		PREFIX wdt: <http://www.wikidata.org/prop/direct/>

		SELECT ?wikidata ?en ?no WHERE {
			?id wdtn:P1015 <https://livedata.bibsys.no/authority/%s>  .
			BIND(STRAFTER(STR(?id), "http://www.wikidata.org/entity/") AS ?wikidata)

			OPTIONAL {
				?wp_no schema:about ?id ; schema:isPartOf <https://no.wikipedia.org/> .
				BIND(STRAFTER(STR(?wp_no), "https://no.wikipedia.org/wiki/") AS ?no)
			}
			OPTIONAL {
				?wp_en schema:about ?id ; schema:isPartOf <https://en.wikipedia.org/> .
				BIND(STRAFTER(STR(?wp_en), "https://en.wikipedia.org/wiki/") AS ?en)
			}
		}
	`
	res, err := repo.Query(fmt.Sprintf(sparqlQ, bibsysID))
	if err != nil {
		return err
	}

	if len(res.Solutions()) != 1 {
		return nil
	}

	*n++

	const q = `INSERT OR IGNORE INTO link (resource_id, type, id) VALUES (?, ?, ?)`

	for k, v := range res.Solutions()[0] {
		linkType := k
		if linkType != "wikidata" {
			linkType = "wikipedia/" + k
		}

		if err := sqlitex.Exec(conn, q, nil, resourceID, linkType, v.String()); err != nil {
			return err
		}
	}

	return nil
}

type HarvestWikipediaSummaries struct {
	DB *sqlitex.Pool
}

func (h *HarvestWikipediaSummaries) Name() string {
	return "harvest_wikipedia_summaries"
}

func (h *HarvestWikipediaSummaries) Run(ctx context.Context, w io.Writer) error {
	conn := h.DB.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer h.DB.Put(conn)

	const q = `
		SELECT l.resource_id, l.type, l.id FROM link l
			LEFT JOIN resource_text rt ON (rt.resource_id = l.resource_id AND rt.source IN('wikipedia/no','wikipedia/no'))
		WHERE l.type IN ('wikipedia/no', 'wikipedia/en') AND rt.id IS NULL
	`

	c := 0
	n := 0
	fn := func(stmt *sqlite.Stmt) error {
		return wikipediaSummary(
			conn,
			stmt.ColumnText(0),
			strings.TrimPrefix(stmt.ColumnText(1), "wikipedia/"),
			stmt.ColumnText(2),
			&c,
			&n,
		)
	}

	if err := sqlitex.Exec(conn, q, fn); err != nil {
		return err
	}

	fmt.Fprintf(w, "Done. Of %d candidates without wikipeia summaries added %d\n", c, n)

	return nil
}

type wpSummaryResults struct {
	WikibaseItem      string `json:"wikibase_item"`
	Description       string `json:"description"`
	DescriptionSource string `json:"description_source"`
	Extract           string `json:"extract"`
	ExtractHTML       string `json:"extract_html"`
}

func wikipediaSummary(conn *sqlite.Conn, resourceID, wpLang, wpID string, c, n *int) error {
	*c++
	url := fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1/page/summary/%s?redirect=true", wpLang, wpID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "hei@knakk.no")
	req.Header.Set("Accept", "application/json; charset=utf-9")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err // TODO: annotate
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("querying %s.wikipedia.org/api/rest_v1/page/summary/%s: http status %d", wpLang, wpID, resp.StatusCode)
	}

	var res wpSummaryResults
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err // TODO: annotate
	}
	if res.Extract == "" {
		return nil
	}

	const q = `
		INSERT INTO resource_text (resource_id, text, source, source_url, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	if err := sqlitex.Exec(
		conn,
		q,
		nil,
		resourceID,
		res.Extract,
		"wikipedia/"+wpLang,
		fmt.Sprintf("https://%s.wikipedia.org/wiki/%s", wpLang, wpID),
		time.Now().Unix(),
	); err != nil {
		return err
	}
	*n++

	return nil
}
