package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"
)

func invalidStructureError(data []byte) error {
	return fmt.Errorf(
		"invalid JSON structure: `%s`\n%s",
		string(data), debug.Stack(),
	)
}

// A definition update tuple
type updateTuple struct {
	id  uint64
	val string
}

// Repository meta information about available updates
type repoMeta struct {
	Count  int
	Hashes []string
}

func (r *repoMeta) UnMarshalJSON(data []byte) (err error) {
	// Decode this nested tuple clusterfuck
	var arr []json.RawMessage
	data, arr, err = traverseTuples(data, 2, 1, 0, 1, 2, 0)
	if err != nil {
		return
	}

	// Finally get the update array
	err = json.Unmarshal(data, &arr)
	if err != nil {
		return
	}

	// Decode updated hashes
	r.Hashes = make([]string, 0, 256)
	r.Count = len(arr)
	for _, data := range arr {
		var update []json.RawMessage
		err = json.Unmarshal(data, &update)
		if err != nil || len(update) < 2 {
			return
		}
		data = update[1]

		var hexStrings []string
		err = json.Unmarshal(data, &hexStrings)
		if err != nil {
			return
		}
		for _, s := range hexStrings {
			r.Hashes = append(r.Hashes, s)
		}
	}

	return
}

// Abstracts hydrus definition and content update types
type hydrusUpdate interface {
	// Parse JSON into structured data
	Parse([]json.RawMessage) error

	// Commit a parsed update to the database. Returns a slice of SHA256 hashes
	// of affected files.
	Commit() ([][32]byte, error)
}

// Update of int -> hash and int -> tag mappings
type definitionUpdate struct {
	tags   map[uint64][]byte
	hashes map[uint64][32]byte
}

func (u *definitionUpdate) Parse(arr []json.RawMessage) (err error) {
	u.tags = make(map[uint64][]byte, 256)
	u.hashes = make(map[uint64][32]byte, 256)

	data := arr[2]
	err = json.Unmarshal(data, &arr)
	if err != nil {
		return
	}
	if len(arr) < 1 {
		return invalidStructureError(data)
	}

	// Both hash and tag updates are optional. Need to discern payload type.
	for _, data := range arr {
		var typ uint64
		typ, data, err = unpackTypedTuple(data)
		if err != nil {
			return
		}
		switch typ {
		case 0:
			err = u.parseHashes(data)
		case 1:
			err = u.parseTags(data)
		default:
			err = fmt.Errorf("unknown tuple type: %d", typ)
			return
		}
		if err != nil {
			return
		}
	}

	return
}

func (u *definitionUpdate) parseHashes(data []byte) (err error) {
	tuples, err := unpackTupleArray(data)
	if err != nil {
		return
	}
	for _, t := range tuples {
		var hash [32]byte
		_, err = hex.Decode(hash[:], []byte(t.val))
		if err != nil {
			return
		}
		u.hashes[t.id] = hash
	}
	return
}

func (u *definitionUpdate) parseTags(data []byte) (err error) {
	tuples, err := unpackTupleArray(data)
	if err != nil {
		return
	}
	for _, t := range tuples {
		u.tags[t.id] = normalizeTag([]byte(t.val))
	}
	return
}

func (u *definitionUpdate) Commit() (modified [][32]byte, err error) {
	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	if len(u.hashes) != 0 {
		buc := tx.Bucket([]byte("hydrus")).Bucket([]byte("hashes"))
		for id, hash := range u.hashes {
			err = buc.Put(encodeUint64(id), hash[:])
			if err != nil {
				return
			}
		}
	}

	if len(u.tags) != 0 {
		buc := tx.Bucket([]byte("hydrus")).Bucket([]byte("tags"))
		for id, tag := range u.tags {
			err = buc.Put(encodeUint64(id), tag)
			if err != nil {
				return
			}
		}
	}

	err = tx.Commit()
	return
}

// Update containing tagID -> hashIDs mappings
type contentUpdate map[uint64][]uint64

func (u *contentUpdate) Parse(arr []json.RawMessage) (err error) {
	*u = make(contentUpdate, 256)

	// Ignore tag removal updates, because Hydrus tags don't map one to one with
	// Hydron tags
	data := arr[2]
	data, arr, err = traverseTuples(data, 0, 1, 0, 1)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &arr)
	if err != nil {
		return
	}

	for _, data := range arr {
		var raw []json.RawMessage
		err = json.Unmarshal(data, &raw)
		if err != nil {
			return
		}
		if len(raw) < 2 {
			return invalidStructureError(data)
		}

		var (
			tag    uint64
			hashes []uint64
		)
		tag, err = parseUint64(raw[0])
		if err != nil {
			return
		}
		err = json.Unmarshal(raw[1], &hashes)
		if err != nil {
			return
		}

		(*u)[tag] = hashes
	}

	return
}

func (u contentUpdate) Commit() (modified [][32]byte, err error) {
	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	var (
		sha256Buc  = tx.Bucket([]byte("sha256"))
		buc        = tx.Bucket([]byte("hydrus"))
		mappingBuc = buc.Bucket([]byte("hash->tags"))
		hashBuc    = buc.Bucket([]byte("hashes"))
		buf        [8]byte
	)
	modified = make([][32]byte, 0, 16)

	// These relations come from an external source and we have no guarantees
	// about their integrity. If either key is nil, simply skip it.
	for tagID, hashIDs := range u {
		for _, id := range hashIDs {
			encodeUint64To(&buf, id)
			hash := hashBuc.Get(buf[:])
			if hash == nil {
				continue
			}

			encodeUint64To(&buf, tagID)
			tags := append(mappingBuc.Get(hash), buf[:]...)
			err = mappingBuc.Put(hash, tags)
			if err != nil {
				return
			}

			if sha256Buc.Get(hash) != nil {
				modified = append(modified, extractSHA256(hash))
			}
		}
	}

	err = tx.Commit()
	return
}

func extractSHA256(buf []byte) (hash [32]byte) {
	for i := range hash {
		hash[i] = buf[i]
	}
	return
}

// Unpack a tuple with a type annotation member
func unpackTypedTuple(data []byte) (uint64, []byte, error) {
	var arr []json.RawMessage
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return 0, nil, err
	}
	if len(arr) < 2 {
		return 0, nil, invalidStructureError(data)
	}

	typ, err := parseUint64(arr[0])
	return typ, arr[1], err
}

func parseUint64(data []byte) (uint64, error) {
	return strconv.ParseUint(string(data), 10, 64)
}

// Unpack an array of tuples found in definition updates
func unpackTupleArray(data []byte) (tuples []updateTuple, err error) {
	var raw [][2]json.RawMessage
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return
	}
	tuples = make([]updateTuple, len(raw))
	for i, r := range raw {
		if len(r) < 2 {
			if len(r) == 1 {
				err = invalidStructureError(r[0])
			} else {
				err = invalidStructureError(nil)
			}
			return
		}

		tuples[i].id, err = parseUint64(r[0])
		if err != nil {
			return
		}
		err = json.Unmarshal(r[1], &tuples[i].val)
		if err != nil {
			return
		}
	}

	return
}

// Encode uint64 to little endian binary buffer
func encodeUint64(i uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, i)
	return buf
}

// Same as encodeUint64, but writes to target buffer without allocations
func encodeUint64To(dest *[8]byte, i uint64) {
	binary.LittleEndian.PutUint64((*dest)[:], i)
}

// Traverse a nested clusterfuck of tuples, until we you reach the part we need.
// indexes specifies the sequence of indexes to traverse by.
func traverseTuples(data []byte, indexes ...int) (
	[]byte, []json.RawMessage, error,
) {
	var arr []json.RawMessage
	for _, i := range indexes {
		err := json.Unmarshal(data, &arr)
		if err != nil {
			return nil, nil, err
		}
		if len(arr) < i+1 {
			return nil, nil, invalidStructureError(data)
		}
		data = arr[i]
	}

	return data, arr, nil
}
