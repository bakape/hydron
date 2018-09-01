package common

import (
	"bytes"
	"strconv"
)

// Able to stingify itself into a bytes.Buffer
type WriterTo interface {
	WriteTo(*bytes.Buffer)
}

// Buffer all of WriterTo into a []byte
func BufferWriter(wt WriterTo) []byte {
	var b bytes.Buffer
	wt.WriteTo(&b)
	return b.Bytes()
}

type TagSource uint8

const (
	User TagSource = iota
	Gelbooru
	Hydrus
)

type TagType uint8

const (
	Undefined TagType = iota
	Author
	Character
	Series
	Rating
	System
)

func (t TagType) WriteTo(w *bytes.Buffer) {
	var s string
	switch t {
	case Undefined:
		s = "undefined"
	case Author:
		s = "author"
	case Character:
		s = "character"
	case Series:
		s = "series"
	case Rating:
		s = "rating"
	case System:
		s = "system"
	}
	w.WriteString(s)
}

// Common fields of Tag and TagFilter
type TagBase struct {
	Type TagType `json:"type"`
	Tag  string  `json:"tag"`
}

// Convert tag to normalized string representation
func (t TagBase) WriteTo(w *bytes.Buffer) {
	if t.Type != Undefined {
		t.Type.WriteTo(w)
		w.WriteByte(':')
	}
	w.WriteString(t.Tag)
}

// Tag of an image
type Tag struct {
	TagBase
	Source TagSource `json:"source"`
	ID     uint64    `json:"-"` // Not defined on freshly-parsed tags
}

// Types of system values to retrieve
type SystemTagType uint8

const (
	Size SystemTagType = iota
	Width
	Height
	Duration
	TagCount
)

func (t SystemTagType) WriterTo(w *bytes.Buffer) {
	var s string
	switch t {
	case Size:
		s = "size"
	case Width:
		s = "width"
	case Height:
		s = "height"
	case Duration:
		s = "duration"
	case TagCount:
		s = "tag_count"
	}
	w.WriteString(s)
}

// Parsed data of system tag
type SystemTag struct {
	Type       SystemTagType
	Comparator string
	Value      uint64
}

func (t SystemTag) WriterTo(w *bytes.Buffer) {
	w.WriteString("system")
	w.WriteByte(':')
	t.Type.WriterTo(w)
	w.WriteString(t.Comparator)
	w.WriteString(strconv.FormatUint(t.Value, 10))
}

// Tag-based filter
type TagFilter struct {
	Negative bool
	TagBase
}

func (t TagFilter) WriterTo(w *bytes.Buffer) {
	if t.Negative {
		w.WriteByte('-')
	}
	t.TagBase.WriteTo(w)
}

// Collection of filters for a query
type FilterSet struct {
	Tag    []TagFilter
	System []SystemTag
}

func (s FilterSet) WriteTo(w *bytes.Buffer) {
	first := true
	for _, t := range s.Tag {
		if first {
			first = false
		} else {
			w.WriteByte(' ')
		}
		t.WriterTo(w)
	}
	for _, t := range s.System {
		if first {
			first = false
		} else {
			w.WriteByte(' ')
		}
		t.WriterTo(w)
	}
}

func (s FilterSet) String() string {
	return string(BufferWriter(s))
}
