package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/bakape/hydron/files"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db     *sql.DB
	sq     squirrel.StatementBuilderType
	driver string
)

func Open() (err error) {
	// Read config file, if any
	var conf struct {
		Driver, Connection string
	}
	buf, err := ioutil.ReadFile(filepath.Join(files.RootPath, "db_conf.json"))
	switch {
	case os.IsNotExist(err):
		err = nil
	case err == nil:
		err = json.Unmarshal(buf, &conf)
		if err != nil {
			return
		}
	default:
		return
	}
	if conf.Driver == "" || conf.Connection == "" {
		conf.Driver = "sqlite3"
		// Shared cache is required for multithreading
		conf.Connection = fmt.Sprintf("file:%s?cache=shared&mode=rwc",
			filepath.Join(files.RootPath, "db.db"))
	}
	driver = conf.Driver

	db, err = sql.Open(conf.Driver, conf.Connection)
	if err != nil {
		return
	}
	sq = squirrel.StatementBuilder.RunWith(squirrel.NewStmtCacheProxy(db))
	switch conf.Driver {
	case "sqlite3":
		// To avoid locking "database locked" errors. Hard limitation of SQLite,
		// when used from multiple threads.
		db.SetMaxOpenConns(1)
	case "postgres":
		sq = sq.PlaceholderFormat(squirrel.Dollar)
	}

	var currentVersion int
	err = sq.Select("val").
		From("main").
		Where("id = 'version'").
		QueryRow().
		Scan(&currentVersion)
	if err != nil {
		if s := err.Error(); strings.HasPrefix(s, "no such table") ||
			s == `pq: relation "main" does not exist` {
			err = nil
		} else {
			return
		}
	}
	return runMigrations(currentVersion, version)
}

// Close database connection
func Close() error {
	return db.Close()
}
