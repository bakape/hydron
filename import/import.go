package imp

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
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
	f       io.ReadSeeker
	size    int
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
				img, err := doImport(req.f, req.size, req.addTags)
				req.res <- response{
					Image: img,
					err:   err,
				}
			}
		}()
	}
}

// Worker function for file importing
func doImport(f io.ReadSeeker, size int, addTags string,
) (r common.Image, err error) {
	// Generate SHA1 hash while streaming file to reduce memory allocations in
	// case of file already existing in the DB. Introduces a small overhead, if
	// the file is not already imported, but that is mostly mitigated by the
	// drive cache.
	var (
		digest = sha1.New()
		buf    [512]byte
		n      int
	)
	for {
		n, err = f.Read(buf[:])
		switch err {
		case nil:
			if n > 0 {
				digest.Write(buf[0:n])
			}
		case io.EOF:
			r.SHA1 = hex.EncodeToString(digest.Sum(nil))
			goto process
		default:
			return
		}
	}

process:
	_, err = f.Seek(0, 0)
	if err != nil {
		return
	}
	srcBuf, err := thumbnailer.ReadInto(thumbnailer.GetBufferCap(size), f)
	if err != nil {
		return
	}
	defer thumbnailer.ReturnBuffer(srcBuf)

	// Check, if not already in the database
	isImported, err := db.IsImported(r.SHA1)
	if err != nil {
		return
	}
	if isImported {
		err = ErrImported
		return
	}

	src, thumb, err := thumbnailer.ProcessBuffer(srcBuf, thumbnailerOpts)
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

	mHash := md5.Sum(srcBuf)
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

	err = writeFile(files.ThumbPath(r.SHA1, thumb.IsPNG), thumb.Data)
	if err != nil {
		return
	}
	err = writeFile(files.SourcePath(r.SHA1, r.Type), srcBuf)
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
openFile:
	f, err := os.OpenFile(path, flags, 0660)
	if err != nil {
		if err = os.MkdirAll(filepath.Dir(path), 0760); err != nil {
			return
		}

		goto openFile
	}
	defer f.Close()

	_, err = f.Write(buf)
	return
}

/*
Attempt to import any readable stream
f: stream to read
size: estimated file size
addTags: Add a list of tags to all images
fetchTags: fetch tags from gelbooru
*/
func ImportFile(f io.ReadSeeker, size int, name string, addTags string, fetchTags bool,
) (r common.Image, err error) {
	ch := make(chan response)
	importFile <- request{f, size, addTags, ch}
	res := <-ch
	r = res.Image
	err = res.err
	if err != nil {
		return
	}

	if len(name) > 128 {
		name = name[0:128]
	}

	err = db.SetName(r.ID, name)
	
	if err != nil {
		return
	}

	r.Name = name

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
