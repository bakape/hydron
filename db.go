package main

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
)

// Boolean flags for a record
const (
	fetchedTags = 1 << iota
	pngThumbnail
)

// Byte offsets withing a record
const (
	offsetBools     = 1
	offsetImportIme = offsetBools + 8*iota
	offsetSize
	offsetWidth
	offsetHeight
	offsetLength
	offsetMD5
	offsetTags = offsetMD5 + 16
)

var (
	db                *bolt.DB
	errRecordNotFound = errors.New("record not found")
)

func openDB() (err error) {
	db, err = bolt.Open(filepath.Join(rootPath, "db.db"), 0600, nil)
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

// Write a record to the database. Optionally specify new tags to write, which
// must be re-indexed.
func writeRecord(kv keyValue, tags [][]byte) (err error) {
	var modifiedTags [][]byte
	if tags != nil {
		// Update in-memory tag index
		old := kv.record.Tags()
		modifiedTags = mergeTagSets(old, tags)

		tagMu.Lock()
		defer tagMu.Unlock()
		unindexFileNoMu(kv.sha1, old)
		for _, t := range tags {
			indexTagNoMu(t, kv.sha1)
		}
		kv.SetTags(tags)
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

	// Write changed record
	err = putRecord(tx, kv.sha1, kv.record)
	if err != nil {
		return
	}

	return tx.Commit()
}

func putRecord(tx *bolt.Tx, id [20]byte, r record) error {
	return tx.Bucket([]byte("images")).Put(id[:], []byte(r))
}

// Sync select tags from memory to disk. Requires `tagMu.Lock()`.
func syncTags(tx *bolt.Tx, tags [][]byte) (err error) {
	buc := tx.Bucket([]byte("tags"))
	for _, t := range tags {
		err = buc.Put(t, encodeTaggedList(tagIndex[string(t)]))
		if err != nil {
			return
		}
	}
	return
}

// Execute a function on all records in the database
func iterateRecords(fn func(k []byte, r record)) error {
	return db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("images")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fn(k, record(v))
		}
		return nil
	})
}

// Add tags to an existing record
func addTags(id [20]byte, tags [][]byte) (err error) {
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
	r.MergeTags(tags)

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

// Retrieve a specific record from the DB
func getRecord(tx *bolt.Tx, id [20]byte) (r record, err error) {
	r = record(tx.Bucket([]byte("images")).Get(id[:]))
	if r == nil {
		err = errRecordNotFound
	}
	return
}

// Remove tags from an existing record
func removeTags(id [20]byte, tags [][]byte) (err error) {
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
	r.SetTags(subtractTags(r.Tags(), tags))
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
func removeFile(id [20]byte) (err error) {
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
	idStr := hex.EncodeToString(id[:])
	paths := [...]string{
		sourcePath(idStr, r.Type()),
		thumbPath(idStr, r.ThumbIsPNG()),
	}
	for _, p := range paths {
		err = os.Remove(p)
		if err != nil {
			return
		}
	}

	// Remove DB record
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

// Remove files from the database by ID from the CLI
func removeFilesCLI(ids []string) error {
	for _, id := range ids {
		sha1, err := stringToSHA1(id)
		if err != nil {
			return err
		}
		err = removeFile(sha1)
		if err != nil {
			return err
		}
	}
	return nil
}
