package main

import (
	"encoding/hex"

	"github.com/bakape/hydron/core"
	"github.com/mailru/easyjson/jwriter"
)

// Encode a Go core.KeuValue into a C Record. minimal specifies, if tags and md5
// fields should be omitted.
func encodeRecord(r core.KeyValue, w *jwriter.Writer, minimal bool) {
	w.RawString(`{"sha1":"`)
	sha1 := hex.EncodeToString(r.SHA1[:])
	w.RawString(sha1)

	w.RawString(`","type":"`)
	w.RawString(core.Extensions[r.Type()])

	w.RawString(`","importTime":`)
	w.Uint64(r.ImportTime())
	w.RawString(`,"size":`)
	w.Uint64(r.Size())
	w.RawString(`,"width":`)
	w.Uint64(r.Width())
	w.RawString(`,"height":`)
	w.Uint64(r.Height())
	w.RawString(`,"length":`)
	w.Uint64(r.Length())

	w.RawString(`,"selected":false`)

	if r.HasThumb() {
		w.RawString(`,"thumbPath":"file:///`)
		w.RawString(core.ThumbPath(sha1, r.ThumbIsPNG()))
		w.RawByte('"')
	}

	if minimal {
		w.RawByte('}')
		return
	}

	w.RawString(`,"sourcePath":"file:///`)
	w.RawString(core.SourcePath(sha1, r.Type()))

	w.RawString(`","md5":"`)
	var buf [32]byte
	md5 := r.MD5()
	hex.Encode(buf[:], md5[:])
	w.Raw(buf[:], nil)

	w.RawString(`","tags":`)
	r.JSONTags(w)
	w.RawByte('}')
}
