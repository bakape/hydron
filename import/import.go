package imp

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/fetch"
	"github.com/bakape/hydron/files"
	"github.com/bakape/hydron/tags"
	"github.com/bakape/thumbnailer"
)

// Common errors
var (
	ErrUnsupportedFile = errors.New("unsupported file type")
	ErrImported        = errors.New("already imported")
	importFile         = make(chan request)
)

var thumbnailerOpts = thumbnailer.Options{
	JPEGQuality: 90,
	ThumbDims: thumbnailer.Dims{
		Width:  200,
		Height: 200,
	},
	AcceptedMimeTypes: common.AllowedMimes,
}

type request struct {
	f       io.Reader
	addTags string
	res     chan<- response
}

type response struct {
	common.Image
	err error
}

// Spawn goroutines to perform parallel thumbnailing and importing
func init() {
	n := runtime.NumCPU() + 1
	for i := 0; i < n; i++ {
		go func() {
			for {
				req := <-importFile
				img, err := doImport(req.f, req.addTags)
				req.res <- response{
					Image: img,
					err:   err,
				}
			}
		}()
	}
}

// Worker function for file importing
func doImport(f io.Reader, addTags string) (r common.Image, err error) {
	srcBuf := bytes.NewBuffer(thumbnailer.GetBuffer())
	defer func() {
		thumbnailer.ReturnBuffer(srcBuf.Bytes())
	}()
	_, err = srcBuf.ReadFrom(f)
	if err != nil {
		return
	}
	sHash := sha1.Sum(srcBuf.Bytes())
	r.SHA1 = hex.EncodeToString(sHash[:])

	// Check, if not already in the database
	isImported, err := db.IsImported(r.SHA1)
	if err != nil {
		return
	}
	if isImported {
		err = ErrImported
		return
	}

	src, thumb, err := thumbnailer.ProcessBuffer(srcBuf.Bytes(),
		thumbnailerOpts)
	switch err {
	case nil:
	case thumbnailer.ErrNoStreams:
		err = ErrUnsupportedFile
		return
	default:
		if _, ok := err.(thumbnailer.UnsupportedMIMEError); ok {
			err = ErrUnsupportedFile
		}
		return
	}
	defer thumbnailer.ReturnBuffer(thumb.Data)

	mHash := md5.Sum(srcBuf.Bytes())
	r = common.Image{
		CompactImage: common.CompactImage{
			Type: common.MimeTypes[src.Mime],
			SHA1: r.SHA1,
			Thumb: common.Thumbnail{
				IsPNG: thumb.IsPNG,
				Dims: common.Dims{
					Width:  uint64(thumb.Width),
					Height: uint64(thumb.Height),
				},
			},
		},
		Dims: common.Dims{
			Width:  uint64(src.Width),
			Height: uint64(src.Height),
		},
		ImportTime: time.Now().Unix(),
		Size:       len(src.Data),
		Duration:   uint64(src.Length / time.Second),
		MD5:        hex.EncodeToString(mHash[:]),
		Tags:       tags.FromString(addTags, common.User),
	}

	// Write files in parallel.
	// Buffered to prevent goroutine leaks on early return.
	errCh := make(chan error, 1)
	go func() {
		errCh <- writeFile(files.ThumbPath(r.SHA1, thumb.IsPNG), thumb.Data)
	}()
	err = writeFile(files.SourcePath(r.SHA1, r.Type), srcBuf.Bytes())
	if err != nil {
		return
	}
	err = <-errCh
	if err != nil {
		return
	}

	r.ID, err = db.WriteImage(r)
	return
}

// Write a file to disk. If a file already exists, because of an interrupted
// write or something, overwrite it.
func writeFile(path string, buf []byte) (err error) {
	const flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	f, err := os.OpenFile(path, flags, 0660)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.Write(buf)
	return
}

/*
Attempt to import any readable stream
f: stream to read
addTags: Add a list of tags to all images
fetchTags: fetch tags from gelbooru
*/
func ImportFile(f io.Reader, addTags string, fetchTags bool) (
	r common.Image, err error,
) {
	ch := make(chan response)
	importFile <- request{f, addTags, ch}
	res := <-ch
	r = res.Image
	err = res.err
	if err != nil {
		return
	}

	if fetchTags && fetch.CanFetchTags(r.Type) {
		var fetched []common.Tag
		fetched, err = fetch.FetchTags(r.MD5)
		if err != nil {
			return
		}
		err = db.AddTags(r.ID, fetched)
		if err != nil {
			return
		}
		r.Tags = append(r.Tags, fetched...)
	}

	return
}
