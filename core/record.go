package core

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"github.com/mailru/easyjson/jwriter"
)

// Boolean flags for a Record
const (
	fetchedTags = 1 << iota
	pngThumbnail
	noThumb
)

// Byte offsets withing a Record
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

type KeyValue struct {
	SHA1 [20]byte
	Record
}

// Marshals to JSON. Minimal flag sets, if only the most crucial data should be
// printed.
func (r KeyValue) ToJSON(w *jwriter.Writer, minimal bool) {
	var arr [40]byte
	buf := arr[:]
	hex.Encode(buf, r.SHA1[:])
	w.RawString(`{"SHA1":"`)
	w.Raw(buf, nil)

	w.RawString(`","type":"`)
	w.RawString(Extensions[r.Type()])
	w.RawString(`","thumbIsPNG":`)
	w.Bool(r.ThumbIsPNG())
	w.RawString(`","noThumb":`)
	w.Bool(!r.HasThumb())

	w.RawString(`,"importTime":`)
	w.Uint64(r.ImportTime())
	w.RawString(`,"size":`)
	w.Uint64(r.Size())
	w.RawString(`,"width":`)
	w.Uint64(r.Width())
	w.RawString(`,"height":`)
	w.Uint64(r.Height())
	w.RawString(`,"length":`)
	w.Uint64(r.Length())

	if minimal {
		w.RawByte('}')
		return
	}

	w.RawString(`,"MD5":"`)
	md5 := r.MD5()
	buf = arr[:32]
	hex.Encode(buf, md5[:])
	w.Raw(buf, nil)

	w.RawString(`","tags":[`)
	for i, t := range r.Tags() {
		if i != 0 {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.Raw(t, nil)
		w.RawByte('"')
	}
	w.RawString(`]}`)
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
type Record []byte

// Clones the current Record. Needed, when you need to store it after the DB
// transaction closes.
func (r Record) Clone() Record {
	clone := make(Record, len(r))
	copy(clone, r)
	return clone
}

func (r Record) Type() FileType {
	return FileType(r[0])
}

func (r Record) setType(t FileType) {
	r[0] = byte(t)
}

func (r Record) ThumbIsPNG() bool {
	return r[offsetBools]&pngThumbnail != 0
}

func (r Record) setPNGThumb() {
	r[offsetBools] |= pngThumbnail
}

func (r Record) setFetchedTags() {
	r[offsetBools] |= fetchedTags
}

func (r Record) HaveFetchedTags() bool {
	return r[offsetBools]&fetchedTags != 0
}

func (r Record) setNoThumb() {
	r[offsetBools] |= noThumb
}

func (r Record) HasThumb() bool {
	return r[offsetBools]&noThumb == 0
}

func (r Record) getUint64(offset int) uint64 {
	return binary.LittleEndian.Uint64(r[offset:])
}

func (r Record) setUint64(offset int, val uint64) {
	binary.LittleEndian.PutUint64(r[offset:], val)
}

func (r Record) ImportTime() uint64 {
	return r.getUint64(offsetImportIme)
}

func (r Record) Size() uint64 {
	return r.getUint64(offsetSize)
}

func (r Record) Width() uint64 {
	return r.getUint64(offsetWidth)
}

func (r Record) Height() uint64 {
	return r.getUint64(offsetHeight)
}

func (r Record) Length() uint64 {
	return r.getUint64(offsetLength)
}

func (r Record) setStats(importTime, size, width, height, length uint64) {
	r.setUint64(offsetImportIme, importTime)
	r.setUint64(offsetSize, size)
	r.setUint64(offsetWidth, width)
	r.setUint64(offsetHeight, height)
	r.setUint64(offsetLength, length)
}

func (r Record) MD5() (MD5 [16]byte) {
	copy(MD5[:], r[offsetMD5:])
	return
}

func (r Record) setMD5(MD5 [16]byte) {
	copy(r[offsetMD5:], MD5[:])
}

func (r Record) Tags() [][]byte {
	raw := r[offsetTags:]
	if len(raw) == 0 {
		return nil
	}

	var (
		tags     = make([][]byte, 0, 32)
		i, start int
		b        byte
	)
	for i, b = range raw {
		switch b {
		case ' ':
			if start != i {
				tags = append(tags, raw[start:i])
			}
			start = i + 1
		}
	}
	tags = append(tags, raw[start:i+1])

	if len(tags) == 0 {
		return nil
	}
	return tags
}

func (r *Record) setTags(t [][]byte) {
	// If one of the tags in t and r somehow share some allocated memory,
	// prepare for infinite memmove. Safer to always replace r with a clone.
	*r = append(r.Clone()[:offsetTags], bytes.Join(t, []byte{' '})...)
}
