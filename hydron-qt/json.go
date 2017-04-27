package main

import (
	"github.com/bakape/hydron/core"
	"github.com/mailru/easyjson/jwriter"
)

// Encodes an multiple core.Reccord into a JSON array
type recordEncoder struct {
	minimal, notFirst bool
	jwriter.Writer
}

func newRecordEncoder(minimal bool) *recordEncoder {
	w := &recordEncoder{
		minimal: minimal,
	}
	w.RawByte('[')
	return w
}

// Write a record to the buffer as JSON
func (w *recordEncoder) write(kv core.KeyValue) {
	if !w.notFirst {
		w.notFirst = true
	} else {
		w.RawByte(',')
	}
	kv.ToJSON(&w.Writer, w.minimal)
}

// Finish and return the resulting JSON as a string
func (w *recordEncoder) string() string {
	w.RawByte(']')
	buf, err := w.BuildBytes()
	if err != nil { // Should never happen
		panic(err)
	}
	return string(buf)
}
