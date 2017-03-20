package hydrus

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"
	"strconv"

	"github.com/bakape/hydron/common"
	"github.com/boltdb/bolt"
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

type tagEntry struct {
	id  uint64
	tag []byte
}

// Committer commits a buffer to one of the two external sorters
type Committer interface {
	Commit(hashes, hashTags io.Writer) error
}

type definitionCommitter struct {
	tags   []tagEntry
	hashes []byte
}

func (c definitionCommitter) Commit(hashes, _ io.Writer) (err error) {
	if len(c.tags) != 0 {
		var tx *bolt.Tx
		tx, err = common.DB.Begin(true)
		if err != nil {
			return
		}
		defer tx.Rollback()

		buc := tx.Bucket([]byte("hydrus")).Bucket([]byte("id->tag"))
		for _, e := range c.tags {
			err = buc.Put(EncodeUint64(e.id), e.tag)
			if err != nil {
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			return
		}
	}

	if len(c.hashes) != 0 {
		_, err = hashes.Write(c.hashes)
		if err != nil {
			return
		}
	}

	return
}

type hashTagCommitter []byte

func (c hashTagCommitter) Commit(_, hashTags io.Writer) error {
	_, err := hashTags.Write(c)
	return err
}

// RepoMeta contains repository meta information about available updates
type RepoMeta struct {
	Count  int
	Hashes []string
}

// UnmarshalJSON implements json.Unmarshaler
func (r *RepoMeta) UnmarshalJSON(data []byte) (err error) {
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

// ParseDefinitionUpdate parses and commits update of int -> hash
// and int -> tag mappings
func parseDefinitionUpdate(arr []json.RawMessage) (
	com definitionCommitter, err error,
) {
	data := arr[2]
	err = json.Unmarshal(data, &arr)
	if err != nil {
		return
	}
	if len(arr) < 1 {
		err = invalidStructureError(data)
		return
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
			com.hashes, err = parseHashUpdate(data)
		case 1:
			com.tags, err = parseTagUpdate(data)
		default:
			err = fmt.Errorf("unknown tuple type: %d", typ)
		}
		if err != nil {
			return
		}
	}

	return
}

func parseHashUpdate(data []byte) (update []byte, err error) {
	tuples, err := unpackTupleArray(data)
	if err != nil {
		return
	}
	update = make([]byte, len(tuples)*40)
	for i, t := range tuples {
		_, err = hex.Decode(update[i*40:], []byte(t.val))
		if err != nil {
			return
		}
		binary.LittleEndian.PutUint64(update[i*40+32:], t.id)
	}
	return
}

func parseTagUpdate(data []byte) (tags []tagEntry, err error) {
	tuples, err := unpackTupleArray(data)
	if err != nil {
		return
	}
	tags = make([]tagEntry, len(tuples))
	for i, t := range tuples {
		tags[i] = tagEntry{
			id:  t.id,
			tag: common.NormalizeTag([]byte(t.val)),
		}
	}
	return
}

// ParseContentUpdate parses and commits update containing
// tagID -> hashIDs mappings
func parseContentUpdate(arr []json.RawMessage) (
	com hashTagCommitter, err error,
) {
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

	var (
		tagID   uint64
		hashIDs []uint64
		buf     [16]byte
	)

	com = make(hashTagCommitter, 0, 1<<10)
	for _, data := range arr {
		var raw []json.RawMessage
		err = json.Unmarshal(data, &raw)
		if err != nil {
			return
		}
		if len(raw) < 2 {
			err = invalidStructureError(data)
			return
		}

		tagID, err = parseUint64(raw[0])
		if err != nil {
			return
		}
		err = json.Unmarshal(raw[1], &hashIDs)
		if err != nil {
			return
		}

		// Encode to binary
		binary.LittleEndian.PutUint64(buf[8:], tagID)
		for _, id := range hashIDs {
			// hash:tag pairs only have keys and are later looked up by prefix
			// search
			binary.LittleEndian.PutUint64(buf[:], id)
			com = append(com, buf[:]...)
		}
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

// EncodeUint64 encodes uint64 to little endian binary buffer
func EncodeUint64(i uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, i)
	return buf
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
