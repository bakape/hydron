package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"time"

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
		decodedTags = splitTagString(tags)
	}
	for i := 0; i < n; i++ {
		go func() {
			for f := range src {
				kv, err := importFile(f, del, decodedTags)
				res <- response{
					keyValue: kv,
					path:     f,
					err:      err,
				}
			}
		}()
	}

	p := progressLogger{
		total:  len(files),
		header: "importing and thumbnailing",
	}
	var toFetch map[[20]byte]record
	if fetchTags {
		toFetch = make(map[[20]byte]record, len(files))
	}
	for range files {
		switch r := <-res; r.err {
		case errUnsupportedFile:
			p.total--
		case errImported:
			p.done++
		case nil:
			if fetchTags {
				switch r.Type() {
				case jpeg, png, gif, webm:
					toFetch[r.SHA1] = r.record
				}
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

func importFile(path string, del bool, tags [][]byte) (
	kv keyValue, err error,
) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

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
		if del {
			err = os.Remove(path)
			return
		}
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
		SHA1:   hash,
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
	kv.SetMD5(md5.Sum(buf))

	// Receive any disk write errors
	for i := 0; i < 2; i++ {
		err = <-errCh
		if err != nil {
			return
		}
	}

	err = writeRecord(kv, tags)
	if err != nil {
		return
	}
	if del {
		err = os.Remove(path)
	}
	return
}
