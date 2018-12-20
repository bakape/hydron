package db

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
)

// Execute all SQL statement strings and return on first error, if any
func execAll(tx *sql.Tx, q ...string) error {
	for _, q := range q {
		if _, err := tx.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// Runs function inside a transaction and handles comminting and rollback on
// error
func InTransaction(fn func(*sql.Tx) error) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}

	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return
	}
	return tx.Commit()
}

// PostgreSQL does not support LastInsertId() and needs some modifications for
// getting it
func getLastID(tx *sql.Tx, q squirrel.InsertBuilder) (id int64, err error) {
	switch driver {
	case "postgres":
		err = q.Suffix("returning id").
			RunWith(tx).
			QueryRow().
			Scan(&id)
		return
	default:
		var res sql.Result
		res, err = q.RunWith(tx).Exec()
		if err != nil {
			return
		}
		id, err = res.LastInsertId()
		return
	}
}
