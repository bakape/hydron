package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"github.com/bakape/hydron/hydrus"
	"github.com/bakape/thumbnailer"
)

var (
	errUnsupportedFile = errors.New("usuported file type")
	errImported        = errors.New("file already imported")

	thumbnailerOpts = thumbnailer.Options{
		JPEGQuality: 90,
		ThumbDims: thumbnailer.Dims{
			Width:  200,
			Height: 200,
		},
	}
)

// Recursively import list of files and/or directories
func importPaths(paths []string, del, fetchTags bool, tags string) (err error) {
	files, err := traverse(paths)
	if err != nil {
		return
	}

	// Parallelize across all cores and report progress in real time
	type response struct {
		keyValue
		path string
		err  error
	}
	src := make(chan string)
	res := make(chan response)
	go func() {
		for _, f := range files {
			src <- f
		}
		close(src)
	}()

	n := runtime.NumCPU() + 1
	var decodedTags [][]byte
	if tags != "" {
		decodedTags = splitTagString(tags, ' ')
	}
	for i := 0; i < n; i++ {
		go func() {
			for path := range src {
				f, err := os.Open(path)
				if err != nil {
					res <- response{
						path: path,
						err:  err,
					}
					continue
				}

				kv, err := importFile(f, decodedTags)
				f.Close()
				if del {
					switch err {
					case nil, errImported:
						err = os.Remove(path)
					}
				}
				res <- response{
					keyValue: kv,
					path:     path,
					err:      err,
				}
			}
		}()
	}

	p := progressLogger{
		total:  len(files),
		header: "importing and thumbnailing",
	}
	var toFetch map[[16]byte]keyValue
	if fetchTags {
		toFetch = make(map[[16]byte]keyValue, len(files))
	}
	for range files {
		switch r := <-res; r.err {
		case errUnsupportedFile:
			p.total--
		case errImported:
			p.done++
		case nil:
			if fetchTags && canFetchTags(r.record) {
				toFetch[r.record.MD5()] = r.keyValue
			}
			p.done++
		default:
			p.total--
			p.printError(fmt.Errorf("%s: %s", r.err, r.path))
		}
		p.print()
	}
	p.close()

	// Fetch tags of any newly imported files
	if len(toFetch) != 0 {
		return fetchFileTags(toFetch)
	}

	return nil
}

// Import file from disk path
func importPath(path string, del bool, tags [][]byte) (kv keyValue, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	kv, err = importFile(f, tags)
	if del {
		switch err {
		case nil, errImported:
			err = os.Remove(path)
		}
	}
	return
}

// Import any readable file
func importFile(f io.Reader, tags [][]byte) (kv keyValue, err error) {
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	hash := sha1.Sum(buf)

	// Check, if not already in the database
	isImported, err := isImported(hash)
	if err != nil {
		return
	} else if isImported {
		err = errImported
		return
	}

	src, thumb, err := thumbnailer.ProcessBuffer(buf, thumbnailerOpts)
	switch err {
	case nil:
	case thumbnailer.ErrNoCoverArt, thumbnailer.ErrNoStreams:
		err = errUnsupportedFile
		return
	default:
		_, ok := err.(thumbnailer.UnsupportedMIMEError)
		if ok {
			err = errUnsupportedFile
		}
		return
	}

	typ := mimeTypes[src.Mime]
	errCh := make(chan error)
	id := hex.EncodeToString(hash[:])
	go func() {
		errCh <- writeFile(sourcePath(id, typ), buf)
	}()
	go func() {
		errCh <- writeFile(thumbPath(id, thumb.IsPNG), thumb.Data)
	}()

	// Create database key-value pair
	kv = keyValue{
		sha1:   hash,
		record: make(record, offsetTags),
	}
	kv.SetType(typ)
	kv.SetStats(
		uint64(time.Now().Unix()),
		uint64(len(buf)),
		uint64(src.Width),
		uint64(src.Height),
		uint64(src.Length/time.Second),
	)
	if thumb.IsPNG {
		kv.SetPNGThumb()
	}

	// Compute hashes concurrently
	sha256Ch := make(chan [32]byte)
	go func() {
		sha256Ch <- sha256.Sum256(buf)
	}()
	kv.SetMD5(md5.Sum(buf))
	kv.sha256 = <-sha256Ch

	// Map tags from Hydrus
	tx, err := db.Begin(false)
	if err != nil {
		return
	}
	defer tx.Rollback()
	if t := hydrus.MapTags(tx, kv.sha256[:]); t != nil {
		tags = mergeTagSets(tags, t)
	}

	// Receive any disk write errors
	for i := 0; i < 2; i++ {
		err = <-errCh
		if err != nil {
			return
		}
	}

	err = writeRecord(kv, tags)
	return
}
