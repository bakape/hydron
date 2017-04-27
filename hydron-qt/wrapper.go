package main

import (
	"strings"

	"github.com/bakape/hydron/core"
	"github.com/mailru/easyjson/jwriter"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/qml"
)

// TODO: Better error reporting

//go:generate qtmoc
type Bridge struct {
	qtcore.QObject

	_ func(id, fileType string) string   `slot:"sourcePath"`
	_ func(id string, isPNG bool) string `slot:"thumbPath"`
	_ func(tags string) string           `slot:"search"`
	_ func(id string) string             `slot:"get"`
	_ func(tagStr string) []string       `slot:"completeTag"`
}

// Bind FFI Go-QML callbacks
func buildBridge(app *qml.QQmlApplicationEngine) {
	bridge := NewBridge(nil)

	bridge.ConnectSourcePath(sourcePath)
	bridge.ConnectThumbPath(thumbPath)
	bridge.ConnectSearch(searchByTags)
	bridge.ConnectGet(getRecord)
	bridge.ConnectCompleteTag(completeTag)

	//make the bridge object accessible inside qml as "go"
	app.RootContext().SetContextProperty("go", bridge)
}

// Return source file path
func sourcePath(id, fileType string) string {
	return "file:///" + core.SourcePath(id, core.RevExtensions[fileType])
}

// Return thumbnail file path
func thumbPath(id string, isPNG bool) string {
	return "file:///" + core.ThumbPath(id, isPNG)
}

// Return record JSON by tag intersection
func searchByTags(tags string) string {
	tags = strings.TrimSpace(tags)
	if tags == "" {
		s, err := encodeAllRecords()
		if err != nil {
			panic(err)
		}
		return s
	}

	matched, err := core.SearchByTags(core.SplitTagString(tags, ' '))
	if err != nil {
		panic(err)
	}

	e := newRecordEncoder(true)
	err = core.MapRecords(matched, func(id [20]byte, r core.Record) {
		e.write(core.KeyValue{
			SHA1:   id,
			Record: r,
		})
	})
	if err != nil {
		panic(err)
	}
	return e.string()
}

// Encode all records in the database as JSON
func encodeAllRecords() (string, error) {
	e := newRecordEncoder(true)
	err := core.IterateRecords(func(k []byte, r core.Record) {
		e.write(core.KeyValue{
			SHA1:   core.ExtractKey(k),
			Record: r,
		})
	})
	return e.string(), err
}

// Retrieve a single record by ID
func getRecord(id string) string {
	sha1, err := core.StringToSHA1(id)
	if err != nil {
		panic(err)
	}
	r, err := core.GetRecord(sha1)
	if err != nil {
		panic(err)
	}

	if r == nil {
		return "null"
	}

	var w jwriter.Writer
	core.KeyValue{
		SHA1:   sha1,
		Record: r,
	}.ToJSON(&w, false)
	buf, err := w.BuildBytes()
	if err != nil {
		panic(err)
	}
	return string(buf)
}

// Complete the last tag in the tag string, if any
func completeTag(tagStr string) []string {
	// Empty field or starting to type a new tag
	if tagStr == "" || tagStr[len(tagStr)-1] == ' ' {
		return nil
	}
	tags := core.SplitTagString(tagStr, ' ')
	if len(tags) == 0 {
		return nil
	}
	return core.CompleteTag(string(tags[len(tags)-1]))
}
