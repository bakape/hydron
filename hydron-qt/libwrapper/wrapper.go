package main

import "C"
import (
	"strings"

	"github.com/bakape/hydron/core"
	"github.com/mailru/easyjson/jwriter"
)

//export searchByTags
// Retrieves records by tag intersection.
// Returns record array as JSON, it's length in bytes and error, if any.
func searchByTags(tags_C *C.char) (*C.char, *C.char) {
	tags := strings.TrimSpace(C.GoString(tags_C))
	if tags == "" {
		return encodeAllRecords()
	}

	matched, err := core.SearchByTags(core.SplitTagString(tags, ' '))
	if err != nil {
		return nil, toCError(err)
	}

	w, comma, flush := encodeArray()
	err = core.MapRecords(matched, func(id [20]byte, r core.Record) {
		comma()
		kv := core.KeyValue{
			SHA1:   id,
			Record: r,
		}
		encodeRecord(kv, w, true)
	})
	if err != nil {
		return nil, toCError(err)
	}

	return flush()
}

// Encodes a Go error, if any, as a C string
func toCError(err error) *C.char {
	if err == nil {
		return nil
	}
	return C.CString(err.Error())
}

// Dumps JSON to a C buffer and returns it with  any error
func toCJSON(w *jwriter.Writer) (*C.char, *C.char) {
	buf, err := w.BuildBytes()
	if err != nil {
		return nil, toCError(err)
	}

	// Coverts to C string without allocating and extra Go string and passing
	// that to C.CString()
	l := len(buf)
	s := C.malloc(C.size_t(l + 1))
	sp := (*[1 << 30]byte)(s)
	copy(sp[:], buf)
	sp[l] = 0
	return (*C.char)(s), nil
}

// Helper for encoding JSON arrays to C strings
func encodeArray() (
	w *jwriter.Writer,
	comma func(),
	flush func() (*C.char, *C.char),
) {
	w = &jwriter.Writer{}
	w.RawByte('[')
	first := true

	return w,
		func() {
			if !first {
				w.RawByte(',')
			} else {
				first = false
			}
		},
		func() (*C.char, *C.char) {
			w.RawByte(']')
			return toCJSON(w)
		}
}

// Encode all records in the database as JSON
func encodeAllRecords() (*C.char, *C.char) {
	w, comma, flush := encodeArray()
	err := core.IterateRecords(func(k []byte, r core.Record) {
		comma()
		kv := core.KeyValue{
			SHA1:   core.ExtractKey(k),
			Record: r,
		}
		encodeRecord(kv, w, true)
	})
	if err != nil {
		return nil, toCError(err)
	}

	return flush()
}

//export getRecord
// Retrieve a single record by ID.
// If no record is found, "null" is returned.
func getRecord(id_C *C.char) (*C.char, *C.char) {
	sha1, err := core.StringToSHA1(C.GoString(id_C))
	if err != nil {
		return nil, toCError(err)
	}

	r, err := core.GetRecord(sha1)
	if err != nil || r == nil {
		return nil, toCError(err)
	}

	var w jwriter.Writer
	kv := core.KeyValue{
		SHA1:   sha1,
		Record: r,
	}
	encodeRecord(kv, &w, false)
	return toCJSON(&w)
}

//export completeTag
// Complete the last tag in the tag string, if any.
// Returns a JSON array of tag strings.
func completeTag(tagStr_C *C.char) (*C.char, *C.char) {
	// Empty field or starting to type a new tag
	tagStr := C.GoString(tagStr_C)
	if tagStr == "" || tagStr[len(tagStr)-1] == ' ' {
		return C.CString("[]"), nil
	}

	w, comma, flush := encodeArray()
	for _, t := range core.CompleteTag(tagStr) {
		comma()
		w.RawByte('"')
		w.RawString(t)
		w.RawByte('"')
	}
	return flush()
}

//export remove
// Remove a file by ID
func remove(id_C *C.char) *C.char {
	id, err := core.StringToSHA1(C.GoString(id_C))
	if err != nil {
		return toCError(err)
	}

	err = core.RemoveFile(id)
	if err != nil {
		return toCError(err)
	}
	return nil
}
