package main

import (
	"bytes"
	"encoding/binary"
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
	db *bolt.DB

	// In memory cache of tags to all applicable images
	tagIndex = make(map[string][][16]byte, 64)
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
	return bytes.Split(r[offsetTags:], []byte{' '})
}

func (r *record) SetTags(t [][]byte) {
	*r = append((*r)[:offsetTags], bytes.Join(t, []byte{' '})...)
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

func writeRecord(kv keyValue) error {
	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("images")).Put(kv.SHA1[:], []byte(kv.record))
	})
}

// Extract a copy of the underlying SHA1 key from a BoltDB key
func extractKey(k []byte) (sha1 [20]byte) {
	for i := range sha1 {
		sha1[i] = k[i]
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
