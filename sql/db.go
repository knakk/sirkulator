package sql

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
)

// Package constants and variables
const (
	memPoolURI  = "file:%s?mode=memory&cache=shared"
	filePoolURI = "file:%s/%s.db?mode=rwc&cache=shared"
	poolSize    = 20 // TODO tune this value, or make it configurable
	poolFlags   = sqlite.SQLITE_OPEN_READWRITE |
		sqlite.SQLITE_OPEN_CREATE |
		sqlite.SQLITE_OPEN_URI |
		sqlite.SQLITE_OPEN_NOMUTEX |
		sqlite.SQLITE_OPEN_SHAREDCACHE

	// TODO the plan was to set PRAGMA foreing_keys=ON in initScript,
	// but this is currently not possible, see:
	// https://github.com/crawshaw/sqlite/issues/131
	// Instead for now enforcing this as a build constraint in crawshaw.io/sqlite fork
	initScript = ``
)

//go:embed assets
var assets embed.FS

// database is a database with known schema.
type database int

// databases
const (
	unknown database = iota
	dbMain
	dbOAI
	dbFiles
	// Sirkultor
	// Users
)

func (d database) String() string {
	if d > 4 || d < 0 {
		d = 0 // "unknown"
	}
	return [...]string{"unknown", "main", "oai", "files"}[d]
}

// OpenMem opens and initalizes a new in-memory main database, and initializes and
// attaches the other databases (oai, files) to main.
func OpenMem() (*sqlitex.Pool, error) {
	pool, err := sqlitex.OpenInit(context.Background(), fmt.Sprintf(memPoolURI, dbMain), poolFlags, poolSize, initScript)
	if err != nil {
		return nil, fmt.Errorf("sql.OpenMem: %w", err)
	}

	return initPool(pool, "")
}

func attachTo(pool *sqlitex.Pool, dir string, poolSize int, dbs ...database) error {
	var conns []*sqlite.Conn
	defer func() {
		for _, conn := range conns {
			pool.Put(conn)
		}
	}()
	for i := 0; i < poolSize; i++ {
		conn := pool.Get(nil)
		if conn == nil {
			return fmt.Errorf("attachTo: cannot get connection %d to attach database", i)
		}
		conns = append(conns, conn)

		for _, db := range dbs {
			file := fmt.Sprintf(filePoolURI, dir, db)
			if dir == "" { // memory-db
				file = fmt.Sprintf(memPoolURI, db)
			}

			stmt, _, err := conn.PrepareTransient("ATTACH DATABASE $file AS $db;")
			if err != nil {
				return fmt.Errorf("attachTo %v: %w", db, err)
			}
			stmt.SetText("$file", file)
			stmt.SetText("$db", db.String())
			_, err = stmt.Step()
			stmt.Finalize()
			if err != nil {
				return fmt.Errorf("attachTo %v: %w", db, err)
			}

			if i == 0 {
				// Run initSchema once per DB
				if err := initSchema(conn, db); err != nil {
					return fmt.Errorf("attachTo %v: %w", db, err)
				}
			}
		}
	}
	return nil
}

// OpenAt opens the and initializes databases at the given directory, creating
// the directory if it doesn't exist. All supplementary databases (oai, files)
// are attached to main.
func OpenAt(dir string) (*sqlitex.Pool, error) {

	// Create directory if it doesn't exist.
	dir = strings.TrimSuffix(dir, "/")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("sql.OpenAt(%s): MkdirAll: %w", dir, err)
	}

	uri := fmt.Sprintf(filePoolURI, dir, "main")
	pool, err := sqlitex.OpenInit(context.Background(), uri, poolFlags, poolSize, initScript)
	if err != nil {
		return nil, fmt.Errorf("sql.OpenAt(%s): OpenInit: %w", uri, err)
	}

	return initPool(pool, dir)
}

func initPool(pool *sqlitex.Pool, dir string) (*sqlitex.Pool, error) {
	if err := attachTo(pool, dir, poolSize, dbOAI, dbFiles); err != nil {
		return pool, fmt.Errorf("sql.OpenAt: %w", err)
	}

	conn := pool.Get(context.Background())
	defer pool.Put(conn)

	// Initialize main db with SQL schema
	if err := initSchema(conn, dbMain); err != nil {
		return nil, fmt.Errorf("sql.OpenAt(%s): %w", dir, err)
	}

	return pool, nil
}

// initSchema initializes the conn with db schema and runs migrations
func initSchema(conn *sqlite.Conn, db database) error {
	version, err := schemaVersion(conn, db)
	if err != nil {
		return fmt.Errorf("initSchema: %w", err)
	}

	if version == 0 {
		schema, err := assets.ReadFile("assets/" + db.String() + "/schema.sql")
		if err != nil {
			return fmt.Errorf("initSchema: read SQL file %w", err)
		}

		if err := sqlitex.ExecScript(conn, string(schema)); err != nil {
			return fmt.Errorf("initSchema: exec SQL schema: %w", err)
		}
	}

	return runMigrations(conn, db)
}

func runMigrations(conn *sqlite.Conn, db database) error {
	path := "assets/" + db.String() + "/migrations"
	dir, err := assets.ReadDir(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("runMigrations: ReadDir: %w", err)
	}

	version, err := schemaVersion(conn, db)
	if err != nil {
		return fmt.Errorf("runMigrations: %w", err)
	}

	for _, file := range dir { // directory entries are sorted by embed package
		if file.IsDir() {
			continue
		}
		d, err := strconv.Atoi(strings.Replace(file.Name(), ".sql", "", 1))
		if err != nil {
			// silently ignore file names not named \d{4}.sql
			continue
		}
		if d == version {
			migration, err := assets.ReadFile(path + "/" + file.Name())
			if err != nil {
				// TODO pool.Close()
				return fmt.Errorf("runMigrations: read migration file: %w", err)
			}
			if err := sqlitex.ExecScript(conn, string(migration)); err != nil {
				// TODO pool.Close()
				return fmt.Errorf("runMigrations: exec migration sql: %w", err)
			}
			version++
		}
	}
	return nil
}

func schemaVersion(conn *sqlite.Conn, db database) (int, error) {
	var v int
	fn := func(stmt *sqlite.Stmt) error {
		v = int(stmt.ColumnInt64(0))
		return nil
	}
	// TODO properly build query string
	if err := sqlitex.ExecTransient(conn, "PRAGMA "+db.String()+".user_version;", fn); err != nil {
		return 0, fmt.Errorf("schemaVersion: %w", err)
	}
	return v, nil
}
