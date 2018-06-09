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

// // Execute a readonly function on all records in the database
// func IterateRecords(fn func(k []byte, r Record)) error {
// 	return db.View(func(tx *bolt.Tx) error {
// 		c := tx.Bucket([]byte("images")).Cursor()
// 		for k, v := c.First(); k != nil; k, v = c.Next() {
// 			fn(k, Record(v))
// 		}
// 		return nil
// 	})
// }

// // Retrieves any records that with the provided ids and runs fn on them. Record
// // is only valid, while MapRecords is executing.
// func MapRecords(ids map[[20]byte]bool, fn func(id [20]byte, r Record)) error {
// 	// Sort keys by ID for more sequential and faster reads
// 	sorted := make(sha1Sorter, len(ids))
// 	i := 0
// 	for id := range ids {
// 		sorted[i] = id
// 		i++
// 	}
// 	sort.Sort(sorted)

// 	tx, err := db.Begin(false)
// 	if err != nil {
// 		return err
// 	}

// 	buc := tx.Bucket([]byte("images"))
// 	for _, id := range sorted {
// 		v := buc.Get(id[:])
// 		if v != nil {
// 			fn(id, Record(v))
// 		}
// 	}

// 	return tx.Rollback()
// }

// // Retrieves a single record by ID from the database. r = nil, if no record
// // found.
// func GetRecord(id [20]byte) (r Record, err error) {
// 	err = db.View(func(tx *bolt.Tx) error {
// 		r = Record(tx.Bucket([]byte("images")).Get(id[:])).Clone()
// 		return nil
// 	})
// 	return
// }

// // Remove tags from an existing Record
// func RemoveTags(id [20]byte, tags [][]byte) (err error) {
// 	tx, err := db.Begin(true)
// 	if err != nil {
// 		return
// 	}
// 	defer tx.Rollback()

// 	r, err := getRecord(tx, id)
// 	if err != nil {
// 		return
// 	}

// 	r.setTags(subtractTags(r.Tags(), tags))

// 	err = putRecord(tx, id, r)
// 	if err != nil {
// 		return
// 	}

// 	err = tx.Commit()
// 	return
// }
