package sql

import (
	"encoding/json"
	"fmt"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator"
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

//func GetOAIRecord(conn *sqlite.Conn, source,id string) (oai.Record, error) {}
