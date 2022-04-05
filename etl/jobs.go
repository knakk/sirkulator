package etl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator/marc"
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

	urnID := strings.TrimPrefix(res.Embedded.Items[0].Metadata.Identifiers.Urn, "URN:NBN:no-nb_digibok_")
	if urnID == "" {
		return nil
	}

	const q = `INSERT OR IGNORE INTO link (resource_id, type, id) VALUES (?, ?, ?)`

	if err := sqlitex.Exec(conn, q, nil, resourceID, linkType, urnID); err != nil {
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
