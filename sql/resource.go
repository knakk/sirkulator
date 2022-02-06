package sql

import (
	"encoding/json"
	"fmt"

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

func GetResource(conn *sqlite.Conn, t sirkulator.ResourceType, id string) (sirkulator.Resource, error) {
	var res sirkulator.Resource
	const q = "SELECT id, label, data FROM resource WHERE type=? AND id=?"
	if err := sqlitex.Exec(conn, q, readData(&res, t), t.String(), id); err != nil {
		return res, fmt.Errorf("sql.GetResource(%s, %s): %w", t.String(), id, err)
	}
	// TODO NotFound error; readData fn should set found bool
	return res, nil
}

//func GetOAIRecord(conn *sqlite.Conn, source,id string) (oai.Record, error) {}
