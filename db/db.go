package db

import (
	"database/sql"
	"path/filepath"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/bakape/hydron/files"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
	sq squirrel.StatementBuilderType
)

func Open() (err error) {
	db, err = sql.Open("sqlite3", filepath.Join(files.RootPath, "db.db"))
	if err != nil {
		return
	}
	// To avoid locking "database locked" errors. Hard limitation of SQLite,
	// when used from multiple threads.
	db.SetMaxOpenConns(1)
	sq = squirrel.StatementBuilder.RunWith(squirrel.NewStmtCacheProxy(db))

	var currentVersion int
	err = sq.Select("val").
		From("main").
		Where("id = 'version'").
		QueryRow().
		Scan(&currentVersion)
	if err != nil {
		if strings.HasPrefix(err.Error(), "no such table") {
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
