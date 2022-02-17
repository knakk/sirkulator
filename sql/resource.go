package sql

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/marc"
)

func readResource(res *sirkulator.Resource, t sirkulator.ResourceType) func(stmt *sqlite.Stmt) error {
	return func(stmt *sqlite.Stmt) error {
		res.Type = t
		res.ID = stmt.ColumnText(0)
		res.Label = stmt.ColumnText(1)
		if err := json.Unmarshal([]byte(stmt.ColumnText(2)), res.Data); err != nil {
			return err
		}
		res.CreatedAt = time.Unix(stmt.ColumnInt64(3), 0)
		res.UpdatedAt = time.Unix(stmt.ColumnInt64(4), 0)
		return nil
	}
}

func readData(res *sirkulator.Resource, t sirkulator.ResourceType) func(stmt *sqlite.Stmt) error {
	// TODO try rewrite with generics
	switch t {
	case sirkulator.TypePerson:
		res.Data = &sirkulator.Person{}
		return readResource(res, t)
	case sirkulator.TypePublication:
		res.Data = &sirkulator.Publication{}
		return readResource(res, t)
	default:
		panic("sql.GetResource: readData: TODO")
	}
}

func readLinks(res *sirkulator.Resource) func(stmt *sqlite.Stmt) error {
	return func(stmt *sqlite.Stmt) error {
		res.Links = append(res.Links, [2]string{stmt.ColumnText(0), stmt.ColumnText(1)})
		return nil
	}
}

func GetResource(conn *sqlite.Conn, t sirkulator.ResourceType, id string) (sirkulator.Resource, error) {
	var res sirkulator.Resource

	const qResouce = "SELECT id, label, data, created_at, updated_at FROM resource WHERE type=? AND id=?"
	if err := sqlitex.Exec(conn, qResouce, readData(&res, t), t.String(), id); err != nil {
		return res, fmt.Errorf("sql.GetResource(%s, %s): %w", t.String(), id, err)
	}
	if res.ID == "" {
		return res, sirkulator.ErrNotFound
	}

	const qLinks = "SELECT type, id FROM link WHERE resource_id=?"
	if err := sqlitex.Exec(conn, qLinks, readLinks(&res), id); err != nil {
		return res, fmt.Errorf("sql.GetResource(%s, %s): %w", t.String(), id, err)
	}

	return res, nil
}

// TODO orderBy year|label orderAsc bool
func GetAgentContributions(conn *sqlite.Conn, id string) ([]sirkulator.AgentContribution, error) {
	var res []sirkulator.AgentContribution

	const q = `
	SELECT
		r.from_id,
		resource.type as res_type,
		resource.label as res_label,
		json_extract(resource.data, '$.year') AS year,
		GROUP_CONCAT(json_extract(r.data, '$.role')) as role
	FROM
		relation r
		JOIN resource ON (from_id=resource.id)
	WHERE
		r.type='has_contributor'
	AND to_id=?
	GROUP BY r.from_id
	ORDER BY year DESC`

	fn := func(stmt *sqlite.Stmt) error {
		c := sirkulator.AgentContribution{}
		c.ID = stmt.ColumnText(0)
		c.Type = sirkulator.ParseResourceType(stmt.ColumnText(1))
		c.Label = stmt.ColumnText(2)
		c.Year = stmt.ColumnInt(3)
		for _, role := range strings.Split(stmt.ColumnText(4), ",") {
			rel, err := marc.ParseRelator(role)
			if err == nil {
				c.Roles = append(c.Roles, rel)
			}
		}
		res = append(res, c)
		return nil
	}

	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return res, fmt.Errorf("sql.GetAgentContributions(%q): %w", id, err)
	}

	return res, nil
}

func GetPublcationContributors(conn *sqlite.Conn, id string) ([]sirkulator.PublicationContribution, error) {
	var res []sirkulator.PublicationContribution

	const q = `
	SELECT
		r.to_id,
		resource.type as res_type,
		resource.label as res_label,
		GROUP_CONCAT(json_extract(r.data, '$.role')) as role
	FROM
		relation r
		JOIN resource ON (to_id=resource.id)
	WHERE
		r.type='has_contributor'
	AND from_id=?
	GROUP BY r.to_id`

	fn := func(stmt *sqlite.Stmt) error {
		c := sirkulator.PublicationContribution{}
		c.Agent.ID = stmt.ColumnText(0)
		c.Agent.Type = sirkulator.ParseResourceType(stmt.ColumnText(1))
		c.Agent.Label = stmt.ColumnText(2)
		for _, role := range strings.Split(stmt.ColumnText(3), ",") {
			rel, err := marc.ParseRelator(role)
			if err == nil {
				c.Roles = append(c.Roles, rel)
			}
		}
		res = append(res, c)
		return nil
	}

	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return res, fmt.Errorf("sql.GetPublicationContributions(%q): %w", id, err)
	}

	return res, nil
}

func GetImage(conn *sqlite.Conn, id string) (*sirkulator.Image, error) {
	var img sirkulator.Image
	const q = "SELECT type, width, height FROM files.image WHERE id=?"
	fn := func(stmt *sqlite.Stmt) error {
		img.ID = id
		img.Type = stmt.ColumnText(0)
		img.Width = stmt.ColumnInt(1)
		img.Height = stmt.ColumnInt(2)
		return nil
	}

	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return nil, fmt.Errorf("sql.GetImage(%q): %w", id, err)
	}
	if img.ID == "" {
		return nil, sirkulator.ErrNotFound
	}

	return &img, nil
}

//func GetOAIRecord(conn *sqlite.Conn, source,id string) (oai.Record, error) {}
