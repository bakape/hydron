package main

// #cgo pkg-config: GraphicsMagick
// #cgo CFLAGS: -std=c11 -D_POSIX_C_SOURCE
// #include "init.h"
// #include "thumbnailer.h"
// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
)

var (
	errTooWide             = errors.New("image too wide") // No such thing
	errTooTall             = errors.New("image too tall")
	errThumbnailingUnknown = errors.New("unknown thumbnailing error")
)

func init() {
	C.magickInit()
}

// UnloadGM safely unloads the GraphicsMagick runtime
func UnloadGM() {
	C.DestroyMagick()
}

// processImage generates a thumbnail from a source image buffer. Returns the
// generated thumbnail's buffer, the source images and thumbnail's dimensions
// if a the generated thumbnail is a PNG image and error, if any.
func processImage(buf []byte) ([]byte, bool, error) {
	cBuf := C.CBytes(buf)
	defer C.free(cBuf)

	ex := &C.ExceptionInfo{}
	defer C.DestroyExceptionInfo(ex)

	var thumb C.struct_Thumbnail
	errCode := C.thumbnail(cBuf, C.size_t(len(buf)), &thumb, ex)
	defer func() {
		if thumb.buf != nil {
			C.free(thumb.buf)
		}
	}()
	var err error
	if ex.reason != nil {
		if ex.severity >= 400 {
			err = extractError(ex)
		}
	} else {
		switch errCode {
		case 0:
		case 1:
			err = errThumbnailingUnknown
		}
	}
	if err != nil {
		return nil, false, err
	}

	return C.GoBytes(thumb.buf, C.int(thumb.size)), bool(thumb.isPNG), nil
}

func extractError(ex *C.ExceptionInfo) error {
	r := C.GoString(ex.reason)
	d := C.GoString(ex.description)
	return fmt.Errorf(`thumbnailer: %s: %s`, r, d)
}
