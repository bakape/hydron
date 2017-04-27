// Package core contains core functionality common to both the CLI and GUI
// interface
package core

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"time"

	"sort"

	"github.com/boltdb/bolt"
)

var (
	db                *bolt.DB
	errRecordNotFound = errors.New("Record not found")
)

func openDB() (err error) {
	db, err = bolt.Open(filepath.Join(rootPath, "db.db"), 0600, &bolt.Options{
		Timeout: time.Second * 5,
	})
	if err != nil {
		return
	}

	// Init all buckets, if not created
	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	for _, b := range [...]string{"tags", "images"} {
		_, err = tx.CreateBucketIfNotExists([]byte(b))
		if err != nil {
			return
		}
	}

	// Load all tags into memory
	c := tx.Bucket([]byte("tags")).Cursor()
	for tag, files := c.First(); tag != nil; tag, files = c.Next() {
		for _, f := range decodeTaggedList(files) {
			indexTagNoMu(tag, f)
		}
	}

	return tx.Commit()
}

// Returns if an image is already imported
func isImported(id [20]byte) (bool, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return false, err
	}
	return tx.Bucket([]byte("images")).Get(id[:]) != nil, tx.Rollback()
}

// Write a Record to the database. Optionally specify new tags to write, which
// must be re-indexed.
func writeRecord(kv KeyValue, tags [][]byte) (err error) {
	var modifiedTags [][]byte
	if tags != nil {
		// Update in-memory tag index
		old := kv.Record.Tags()
		modifiedTags = mergeTagSets(old, tags)

		tagMu.Lock()
		defer tagMu.Unlock()
		unindexFileNoMu(kv.SHA1, old)
		for _, t := range tags {
			indexTagNoMu(t, kv.SHA1)
		}
		kv.setTags(tags)
	}

	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	// Write changed tags to the database
	err = syncTags(tx, modifiedTags)
	if err != nil {
		return
	}

	// Write changed Record
	err = putRecord(tx, kv.SHA1, kv.Record)
	if err != nil {
		return
	}

	return tx.Commit()
}

func putRecord(tx *bolt.Tx, id [20]byte, r Record) error {
	err := tx.Bucket([]byte("images")).Put(id[:], []byte(r))
	if err != nil {
		err = wrapError(err, "db: putting Record %v", id)
	}
	return err
}

// Sync select tags from memory to disk. Requires `tagMu.Lock()`.
func syncTags(tx *bolt.Tx, tags [][]byte) (err error) {
	buc := tx.Bucket([]byte("tags"))
	for _, t := range tags {
		if len(t) == 0 {
			continue
		}
		err = buc.Put(t, encodeTaggedList(tagIndex[string(t)]))
		if err != nil {
			err = wrapError(err, "db: syncing tag %s", t)
			return
		}
	}
	return
}

// Execute a readonly function on all records in the database
func IterateRecords(fn func(k []byte, r Record)) error {
	return db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("images")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fn(k, Record(v))
		}
		return nil
	})
}

// Retrieves any records that with the provided ids and runs fn on them. Record
// is only valid, while MapRecords is executing.
func MapRecords(ids map[[20]byte]bool, fn func(id [20]byte, r Record)) error {
	// Sort keys by ID for more sequential and faster reads
	sorted := make(sha1Sorter, len(ids))
	i := 0
	for id := range ids {
		sorted[i] = id
		i++
	}
	sort.Sort(sorted)

	tx, err := db.Begin(false)
	if err != nil {
		return err
	}

	buc := tx.Bucket([]byte("images"))
	for _, id := range sorted {
		v := buc.Get(id[:])
		if v != nil {
			fn(id, Record(v))
		}
	}

	return tx.Rollback()
}

// Add tags to an existing Record
func AddTags(id [20]byte, tags [][]byte) (err error) {
	tagMu.Lock()
	defer tagMu.Unlock()

	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	r, err := getRecord(tx, id)
	if err != nil {
		return
	}
	for _, t := range tags {
		indexTagNoMu(t, id)
	}
	r.setTags(mergeTagSets(r.Tags(), tags))

	err = putRecord(tx, id, r)
	if err != nil {
		return
	}
	err = syncTags(tx, tags)
	if err != nil {
		return
	}

	return tx.Commit()
}

// Retrieve a specific Record from the DB
func getRecord(tx *bolt.Tx, id [20]byte) (r Record, err error) {
	r = Record(tx.Bucket([]byte("images")).Get(id[:]))
	if r == nil {
		err = errRecordNotFound
	}
	return
}

// Retrieves a single record by ID from the database. r = nil, if nor record
// found.
func GetRecord(id [20]byte) (r Record, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		r = Record(tx.Bucket([]byte("images")).Get(id[:])).Clone()
		return nil
	})
	return
}

// Remove tags from an existing Record
func RemoveTags(id [20]byte, tags [][]byte) (err error) {
	tagMu.Lock()
	defer tagMu.Unlock()

	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	r, err := getRecord(tx, id)
	if err != nil {
		return
	}

	unindexFileNoMu(id, tags)
	r.setTags(subtractTags(r.Tags(), tags))
	err = syncTags(tx, tags)
	if err != nil {
		return
	}

	err = putRecord(tx, id, r)
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}

// Remove a file from the database by ID. Non-existant files are ignored.
func RemoveFile(id [20]byte) (err error) {
	tagMu.Lock()
	defer tagMu.Unlock()

	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	r, err := getRecord(tx, id)
	switch err {
	case nil:
	case errRecordNotFound:
		return nil
	default:
		return
	}

	// Remove files
	var (
		idStr = hex.EncodeToString(id[:])
		paths [2]string
	)
	paths[0] = SourcePath(idStr, r.Type())
	if r.HasThumb() {
		paths[1] = ThumbPath(idStr, r.ThumbIsPNG())
	}
	for _, p := range paths {
		if p == "" {
			continue
		}
		err = os.Remove(p)
		if err != nil {
			return
		}
	}

	// Remove DB Record
	err = tx.Bucket([]byte("images")).Delete(id[:])
	if err != nil {
		return
	}

	// Update tag index
	tags := r.Tags()
	unindexFileNoMu(id, tags)
	err = syncTags(tx, tags)
	if err != nil {
		return
	}

	return tx.Commit()
}
