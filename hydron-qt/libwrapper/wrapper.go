package main

import "C"
import (
	"strings"

	"github.com/bakape/hydron/core"
	"github.com/mailru/easyjson/jwriter"
)

// // Bind FFI Go-QML callbacks
// func buildBridge(app *qml.QQmlApplicationEngine) {
// 	bridge := NewBridge(nil)

// 	bridge.ConnectGet(getRecord)
// 	bridge.ConnectCompleteTag(completeTag)

// 	//make the bridge object accessible inside qml as "go"
// 	app.RootContext().SetContextProperty("go", bridge)
// }

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

	var w jwriter.Writer
	w.RawByte('[')
	first := true
	err = core.MapRecords(matched, func(id [20]byte, r core.Record) {
		if !first {
			w.RawByte(',')
		}
		first = false

		kv := core.KeyValue{
			SHA1:   id,
			Record: r,
		}
		encodeRecord(kv, &w, true)
	})
	if err != nil {
		return nil, toCError(err)
	}
	w.RawByte(']')

	return toCJSON(&w)
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
	return C.CString(string(buf)), nil
}

// Encode all records in the database as JSON
func encodeAllRecords() (*C.char, *C.char) {
	var w jwriter.Writer
	w.RawByte('[')
	first := true
	err := core.IterateRecords(func(k []byte, r core.Record) {
		if !first {
			w.RawByte(',')
		}
		first = false

		kv := core.KeyValue{
			SHA1:   core.ExtractKey(k),
			Record: r,
		}
		encodeRecord(kv, &w, true)
	})
	if err != nil {
		return nil, toCError(err)
	}
	w.RawByte(']')

	return toCJSON(&w)
}

// //export getRecord
// // Retrieve a single record by ID.
// // If no record is found, "null" is returned.
// func getRecord(id_C *C.char) (C.Record, *C.char) {
// 	sha1, err := core.StringToSHA1(C.GoString(id_C))
// 	if err != nil {
// 		return C.Record{}, toCError(err)
// 	}
// 	r, err := core.GetRecord(sha1)
// 	if err != nil || r == nil {
// 		return C.Record{}, toCError(err)
// 	}
// 	kv := core.KeyValue{
// 		SHA1:   sha1,
// 		Record: r,
// 	}
// 	return encodeRecord(kv, false), nil
// }

// //export completeTag
// // Complete the last tag in the tag string, if any.
// // Returns an array of tag strings and their length.
// func completeTag(tagStr_C *C.char) C.Tags {
// 	// Empty field or starting to type a new tag
// 	tagStr := C.GoString(tagStr_C)
// 	if tagStr == "" || tagStr[len(tagStr)-1] == ' ' {
// 		return C.Tags{}
// 	}
// 	return encodeTags(core.SplitTagString(tagStr, ' '))
// }
