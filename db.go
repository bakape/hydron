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
	MD5 [16]byte
	record
}

/*
Record of a stored image in the DB. Serialized to binary in the following
format:

byte - file type of source image
byte - file type of thumbnail
byte - boolean flags
uint64 - import Unix time
[20]byte - SHA1 hash
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

func (r record) GetType() fileType {
	return fileType(r[0])
}

func (r record) SetType(t fileType) {
	r[0] = byte(t)
}

func (r record) GetThumbType() fileType {
	return fileType(r[1])
}

func (r record) SetThumbType(t fileType) {
	r[1] = byte(t)
}

func (r record) SetFetchedTags() {
	r[2] |= fetchedTags
}

func (r record) GetFetchedTags() bool {
	return r[2]&fetchedTags != 0
}

func (r record) GetImportTime() uint64 {
	b := []byte(r[3:11])
	return binary.LittleEndian.Uint64(b)
}

func (r record) SetImportTime(t uint64) {
	b := []byte(r[3:])
	binary.LittleEndian.PutUint64(b, t)
}

func (r record) GetSHA1() (SHA1 [20]byte) {
	copy(SHA1[:], []byte(r[11:]))
	return
}

func (r record) SetSHA1(SHA1 [20]byte) {
	copy([]byte(r[11:]), SHA1[:])
}

func (r record) GetTags() [][]byte {
	return bytes.Split([]byte(r[31:]), []byte{' '})
}

func (r *record) SetTags(t [][]byte) {
	b := bytes.Join(t, []byte{' '})
	*r = append((*r)[:31], record(b)...)
}

// Returns if an image is already imported
func isImported(id [16]byte) (bool, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return false, err
	}
	return tx.Bucket([]byte("images")).Get(id[:]) != nil, tx.Rollback()
}

func writeRecord(kv keyValue) error {
	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("images")).Put(kv.MD5[:], []byte(kv.record))
	})
}
