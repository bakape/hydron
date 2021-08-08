package imp

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/fetch"
	"github.com/bakape/hydron/files"
	"github.com/bakape/hydron/tags"
	"github.com/bakape/thumbnailer/v2"
	"github.com/chai2010/webp"
)

// Common errors
var (
	ErrUnsupportedFile = errors.New("unsupported file type")
	ErrImported        = errors.New("already imported")
	importFile         = make(chan request)

	// Pool of temp buffers used for hashing
	buf512Pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 512)
		},
	}

	thumbnailerOpts = thumbnailer.Options{
		ThumbDims: thumbnailer.Dims{
			Width:  200,
			Height: 200,
		},
		AcceptedMimeTypes: common.AllowedMimes,
	}
	webpOpts = webp.Options{
		Quality: 90,
	}
)

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
			runtime.LockOSThread()
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

// Hash file to hx string
func hashFile(rs io.ReadSeeker, h hash.Hash) (
	hash string, read int, err error,
) {
	_, err = rs.Seek(0, 0)
	if err != nil {
		return
	}
	buf := buf512Pool.Get().([]byte)
	defer buf512Pool.Put(buf)

	for {
		buf = buf[:512] // Reset slicing

		var n int
		n, err = rs.Read(buf)
		buf = buf[:n]
		read += n
		switch err {
		case nil:
			h.Write(buf)
		case io.EOF:
			err = nil
			hash = hex.EncodeToString(h.Sum(buf))
			return
		default:
			return
		}
	}
}

// Worker function for file importing
func doImport(f io.ReadSeeker, size int, addTags string,
) (
	r common.Image, err error,
) {
	// Generate SHA1 hash while streaming file to reduce memory allocations in
	// case of file already existing in the DB. Introduces a small overhead, if
	// the file is not already imported, but that is mostly mitigated by the
	// drive cache.
	r.SHA1, r.Size, err = hashFile(f, sha1.New())
	if err != nil {
		return
	}

	// Check, if not already in the database
	isImported, err := db.IsImported(r.SHA1)
	if err != nil {
		return
	}
	if isImported {
		err = ErrImported
		return
	}

	src, thumb, err := thumbnailer.Process(f, thumbnailerOpts)
	switch err {
	case nil:
	case thumbnailer.ErrCantThumbnail:
		err = ErrUnsupportedFile
		return
	default:
		if _, ok := err.(thumbnailer.ErrUnsupportedMIME); ok {
			err = ErrUnsupportedFile
		}
		return
	}

	r.MD5, _, err = hashFile(f, md5.New())
	if err != nil {
		return
	}
	bounds := thumb.Bounds()
	r.CompactImage = common.CompactImage{
		Type: common.MimeTypes[src.Mime],
		SHA1: r.SHA1,
		Thumb: common.Dims{
			Width:  uint64(bounds.Dx()),
			Height: uint64(bounds.Dy()),
		},
	}
	r.Dims = common.Dims{
		Width:  uint64(src.Width),
		Height: uint64(src.Height),
	}
	r.ImportTime = time.Now().Unix()
	r.Duration = uint64(src.Length / time.Second)
	r.Tags = tags.FromString(addTags, common.User)

	// Encode thumbnail and dump source file concurrently
	ch := make(chan error)
	go func() {
		ch <- func() (err error) {
			f, err := createFile(files.ThumbPath(r.SHA1))
			if err != nil {
				return
			}
			defer f.Close()

			return webp.Encode(f, thumb, &webpOpts)
		}()
	}()

	_, err = f.Seek(0, 0)
	if err != nil {
		return
	}
	srcDest, err := createFile(files.SourcePath(r.SHA1, r.Type))
	if err != nil {
		return
	}
	defer srcDest.Close()
	_, err = io.Copy(srcDest, f)
	if err != nil {
		return
	}

	err = <-ch
	if err != nil {
		return
	}

	r.ID, err = db.WriteImage(r)
	return
}

// Create a file for wrting an image
func createFile(path string) (f *os.File, err error) {
openFile:
	f, err = os.Create(path)
	if err != nil {
		err = os.MkdirAll(filepath.Dir(path), 0760)
		if err != nil {
			return
		}
		goto openFile
	}
	return
}

/*
Attempt to import any readable stream
f: stream to read
size: estimated file size
addTags: Add a list of tags to all images
fetchTags: fetch tags from danbooru
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

	if len(name) > 200 {
		name = name[0:200]
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
