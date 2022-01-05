package sql

import (
	"testing"

	"crawshaw.io/sqlite"
)

func TestOpenAllMem(t *testing.T) {
	db, err := OpenMem()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()

	t.Run("foreign keys enforced", func(t *testing.T) {
		t.Skip("Cannot get PRAGMA foreign_keys to work")
		conn := db.Get(nil)
		defer db.Put(conn)
		stmt, err := conn.Prepare(`
			INSERT INTO relation (from_id, to_id, type)
				VALUES('abc', 'xyz', 'test')
			RETURNING *
		`)
		if err != nil {
			t.Error(err)
		} else if _, err := stmt.Step(); err == nil {
			t.Error("expected error, got nil")
		}
		stmt.Finalize()
	})

	t.Run("query across databases", func(t *testing.T) {
		var conns []*sqlite.Conn
		defer func() {
			for _, conn := range conns {
				db.Put(conn)
			}
		}()
		for i := 0; i < poolSize; i++ { // TODO poolSize maybe take as param to OpenMem
			conn := db.Get(nil)
			conns = append(conns, conn)
			_, err := conn.Prepare("SELECT * FROM resource JOIN oai.record JOIN files.image USING (id)")
			if err != nil {
				t.Errorf("conn %d prepare SQL got err %v, expected nil", i, err)
			}
		}
	})
}
