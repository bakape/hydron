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

type executor interface {
	Exec(args ...interface{}) (sql.Result, error)
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

type tableScanner interface {
	rowScanner
	Next() bool
	Err() error
	Close() error
}

// Allows easily running squirrel queries with transactions
type transactionalQuery struct {
	tx *sql.Tx
	sq squirrel.Sqlizer
}

func withTransaction(tx *sql.Tx, q squirrel.Sqlizer) transactionalQuery {
	return transactionalQuery{
		tx: tx,
		sq: q,
	}
}

func (t transactionalQuery) Exec() (err error) {
	sql, args, err := t.sq.ToSql()
	if err != nil {
		return
	}
	_, err = t.tx.Exec(sql, args...)
	return
}

func (t transactionalQuery) Query() (ts tableScanner, err error) {
	sql, args, err := t.sq.ToSql()
	if err != nil {
		return
	}
	return t.tx.Query(sql, args...)
}

func (t transactionalQuery) QueryRow() (rs rowScanner, err error) {
	sql, args, err := t.sq.ToSql()
	if err != nil {
		return
	}
	rs = t.tx.QueryRow(sql, args...)
	return
}

// WrapError wraps error types to create compound error chains
func WrapError(text string, err error) error {
	return wrappedError{
		text:  text,
		inner: err,
	}
}

type wrappedError struct {
	text  string
	inner error
}

func (e wrappedError) Error() string {
	text := e.text
	if e.inner != nil {
		text += ": " + e.inner.Error()
	}
	return text
}

// Helper for lazily preparing statement
type preparedStatement struct {
	str   string
	tx    *sql.Tx
	query *sql.Stmt
	err   error
	row   *sql.Row
}

// Lazily prepare a statement for transaction
func lazyPrepare(tx *sql.Tx, query string) preparedStatement {
	return preparedStatement{
		str: query,
		tx:  tx,
	}
}

func (p *preparedStatement) prepare() {
	if p.query != nil || p.err != nil {
		return
	}
	p.query, p.err = p.tx.Prepare(p.str)
}

func (p *preparedStatement) Exec(args ...interface{}) (sql.Result, error) {
	p.prepare()
	if p.err != nil {
		return nil, p.err
	}
	var res sql.Result
	res, p.err = p.query.Exec(args...)
	return res, p.err
}

func (p *preparedStatement) Query(args ...interface{}) (tableScanner, error) {
	p.prepare()
	if p.err != nil {
		return nil, p.err
	}
	var tb tableScanner
	tb, p.err = p.query.Query(args...)
	return tb, p.err
}

func (p *preparedStatement) QueryRow(args ...interface{}) rowScanner {
	p.prepare()
	if p.err != nil {
		return p
	}
	p.row = p.query.QueryRow(args...)
	return p
}

func (p *preparedStatement) Scan(dest ...interface{}) error {
	if p.err != nil {
		return p.err
	}
	return p.row.Scan(dest...)
}
