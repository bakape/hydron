package main

// #include <string.h>
// #include "types.h"
import "C"
import (
	"encoding/hex"
	"unsafe"

	"github.com/bakape/hydron/core"
)

// Encode a Go core.KeuValue into a C Record. minimal specifies, if tags and md5
// fields should be omitted.
func encodeRecord(kv core.KeyValue, minimal bool) C.Record {
	dst := C.Record{
		_type:      uint32(kv.Type()),
		importTime: C.uint64_t(kv.ImportTime()),
		size:       C.uint64_t(kv.Size()),
		width:      C.uint64_t(kv.Width()),
		height:     C.uint64_t(kv.Height()),
		length:     C.uint64_t(kv.Length()),
	}

	// Go can't cast arrays. Need to loop.
	for i := 0; i < 20; i++ {
		dst.sha1[i] = C.char(kv.SHA1[i])
	}

	if kv.HasThumb() {
		path := core.ThumbPath(hex.EncodeToString(kv.SHA1[:]), kv.ThumbIsPNG())
		dst.thumbPath = C.CString("file:///" + path)
	}

	if !minimal {
		path := core.SourcePath(hex.EncodeToString(kv.SHA1[:]), kv.Type())
		dst.sourcePath = C.CString("file:///" + path)

		md5 := kv.MD5()
		for i := 0; i < 16; i++ {
			dst.md5[i] = C.char(md5[i])
		}

		dst.tags = encodeTags(kv.Tags())
	}

	return dst
}

// Encode tag slice to C array
func encodeTags(tags [][]byte) C.Tags {
	l := len(tags)
	if l == 0 {
		return C.Tags{}
	}

	size := unsafe.Sizeof(uintptr(0))
	buf := C.malloc(C.size_t(l))
	for i, t := range tags {
		pos := uintptr(buf) + size*uintptr(i)
		*(**C.char)(unsafe.Pointer(pos)) = C.CString(string(t))
	}
	return C.Tags{
		tags: (**C.char)(buf),
		len:  C.int(l),
	}
}

// Converts a Go Record array to a C Record array
func encodeRecordArray(recs []C.Record) (*C.Record, C.int) {
	l := len(recs)
	if l == 0 {
		return nil, 0
	}

	// The memory layout is the same. Just allocate a copy.
	size := C.size_t(unsafe.Sizeof(C.Record{})) * C.size_t(l)
	buf := C.malloc(size)
	C.memcpy(buf, unsafe.Pointer(&recs[0]), size)
	return (*C.Record)(buf), C.int(l)
}
