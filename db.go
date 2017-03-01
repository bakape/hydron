package main

import (
	"bytes"
	"encoding/binary"
	"errors"
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

type keyValue struct {
	SHA1 [20]byte
	record
}

/*
Record of a stored file in the DB. Serialized to binary in the following format:

byte - file type of source image
byte - boolean flags
uint64 - import Unix time
uint64 - source image size in bytes
uint64 - source image width
uint64 - source image height
uint64 - source image stream length, if a any
[16]byte - MD5 hash
string... - space-delimited list of tags
*/
type record []byte

// Clones the current record. Needed, when you need to store it after the DB
// transaction closes.
func (r record) Clone() record {
	clone := make(record, len(r))
	copy(clone, r)
	return clone
}

func (r record) Type() fileType {
	return fileType(r[0])
}

func (r record) SetType(t fileType) {
	r[0] = byte(t)
}

func (r record) ThumbIsPNG() bool {
	return r[offsetBools]&pngThumbnail != 0
}

func (r record) SetPNGThumb() {
	r[offsetBools] |= pngThumbnail
}

func (r record) SetFetchedTags() {
	r[offsetBools] |= fetchedTags
}

func (r record) HaveFetchedTags() bool {
	return r[offsetBools]&fetchedTags != 0
}

func (r record) getUint64(offset int) uint64 {
	return binary.LittleEndian.Uint64(r[offset:])
}

func (r record) setUint64(offset int, val uint64) {
	binary.LittleEndian.PutUint64(r[offset:], val)
}

func (r record) ImportTime() uint64 {
	return r.getUint64(offsetImportIme)
}

func (r record) Size() uint64 {
	return r.getUint64(offsetSize)
}

func (r record) Width() uint64 {
	return r.getUint64(offsetWidth)
}

func (r record) Height() uint64 {
	return r.getUint64(offsetHeight)
}

func (r record) Length() uint64 {
	return r.getUint64(offsetLength)
}

func (r record) SetStats(importTime, size, width, height, length uint64) {
	r.setUint64(offsetImportIme, importTime)
	r.setUint64(offsetSize, size)
	r.setUint64(offsetWidth, width)
	r.setUint64(offsetHeight, height)
	r.setUint64(offsetLength, length)
}

func (r record) MD5() (MD5 [16]byte) {
	copy(MD5[:], r[offsetMD5:])
	return
}

func (r record) SetMD5(MD5 [16]byte) {
	copy(r[offsetMD5:], MD5[:])
}

func (r record) Tags() [][]byte {
	raw := r[offsetTags:]
	if len(raw) == 0 {
		return nil
	}

	var (
		tags     = make([][]byte, 0, 32)
		i, start int
		b        byte
	)
	for i, b = range raw {
		switch b {
		case ' ':
			if start != i {
				tags = append(tags, raw[start:i])
			}
			start = i + 1
		}
	}
	tags = append(tags, raw[start:i+1])

	if len(tags) == 0 {
		return nil
	}
	return tags
}

func (r *record) SetTags(t [][]byte) {
	// If one of the tags in t and r somehow share some allocated memory,
	// prepare for infinite memmove. Safer to always replace r with a clone.
	*r = append(r.Clone()[:offsetTags], bytes.Join(t, []byte{' '})...)
}

// Merge a set of tags with the existing ones in the record
func (r *record) MergeTags(t [][]byte) {
	r.SetTags(mergeTagSets(r.Tags(), t))
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
		unindexFileNoMu(kv.SHA1, old)
		for _, t := range tags {
			indexTagNoMu(t, kv.SHA1)
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
	err = putRecord(tx, kv.SHA1, kv.record)
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
