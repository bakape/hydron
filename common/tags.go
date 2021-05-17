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
	Danbooru
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
	Meta
	MD5Field
	SHA1Field
	Name
)

var (
	tagTypeStr = [...]string{"undefined", "author", "character", "series",
		"rating", "system", "meta", "md5", "sha1", "name"}
	systemTagStr = [...]string{"size", "width", "height", "duration",
		"tag_count", "type"}
)

func (t TagType) WriteTo(w *bytes.Buffer) {
	w.WriteString(tagTypeStr[int(t)])
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
	Type
)

func (t SystemTagType) WriteTo(w *bytes.Buffer) {
	w.WriteString(systemTagStr[int(t)])
}

// Parsed data of system tag
type SystemTag struct {
	Type       SystemTagType
	Comparator string
	Value      uint64
}

func (t SystemTag) WriteTo(w *bytes.Buffer) {
	w.WriteString("system:")
	t.Type.WriteTo(w)
	w.WriteString(t.Comparator)
	// system:type needs to be converted back to its extension string
	// from its internal enum representation
	if t.Type == Type {
		w.WriteString(Extensions[FileType(t.Value)])
	} else {
		w.WriteString(strconv.FormatUint(t.Value, 10))
	}
}

// Tag-based filter
type TagFilter struct {
	Negative bool
	TagBase
}

func (t TagFilter) WriteTo(w *bytes.Buffer) {
	if t.Negative {
		w.WriteByte('-')
	}
	t.TagBase.WriteTo(w)
}

// Collection of filters for a query
type FilterSet struct {
	Tag       []TagFilter
	System    []SystemTag
	SystemStr []TagBase
}

func (s FilterSet) WriteTo(w *bytes.Buffer) {
	first := true
	write := func(t WriterTo) {
		if first {
			first = false
		} else {
			w.WriteByte(' ')
		}
		t.WriteTo(w)
	}

	for _, t := range s.Tag {
		write(t)
	}
	for _, t := range s.System {
		write(t)
	}
	for _, t := range s.SystemStr {
		write(t)
	}
}

func (s FilterSet) String() string {
	return string(BufferWriter(s))
}
