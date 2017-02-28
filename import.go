package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/bakape/thumbnailer"
)

var (
	errUnsupportedFile = errors.New("usuported file type")

	thumbnailerOpts = thumbnailer.Options{
		JPEGQuality: 90,
		ThumbDims: thumbnailer.Dims{
			Width:  200,
			Height: 200,
		},
	}
)

// Recursively import list of files and/or directories
func importPaths(paths []string, del bool, tags string) (err error) {
	files, err := traverse(paths)
	if err != nil {
		return err
	}

	// Parallelize across all cores and report progress in real time
	type response struct {
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
				res <- response{
					path: f,
					err:  importFile(f, del, decodedTags),
				}
			}
		}()
	}

	p := progressLogger{
		total:  len(files),
		header: "importing and thumbnailing",
	}
	defer p.close()
	for range files {
		switch r := <-res; r.err {
		case errUnsupportedFile:
			p.total--
		case nil:
			p.done++
		default:
			p.total--
			p.printError(fmt.Errorf("%s: %s", r.err, r.path))
		}
		p.print()
	}

	return nil
}

func importFile(path string, del bool, tags [][]byte) (err error) {
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
	if is, err := isImported(hash); err != nil {
		return err
	} else if is {
		if del {
			return os.Remove(path)
		}
		return nil
	}

	src, thumb, err := thumbnailer.ProcessBuffer(buf, thumbnailerOpts)
	if err != nil {
		_, ok := err.(thumbnailer.UnsupportedMIMEError)
		if ok || err == thumbnailer.ErrNoCoverArt {
			return errUnsupportedFile
		}
		return err
	}

	typ := mimeTypes[src.Mime]
	errCh := make(chan error)
	id := hex.EncodeToString(hash[:])
	go func() {
		path := filepath.FromSlash(sourcePath(id, typ))
		errCh <- writeFile(path, buf)
	}()
	go func() {
		// Some files may not have thumbnails. Simply skip and leave the burden
		// of handling missing thumbnails for later code.
		if thumb.Data == nil {
			errCh <- nil
			return
		}
		path := filepath.FromSlash(thumbPath(id, thumb.IsPNG))
		errCh <- writeFile(path, thumb.Data)
	}()

	// Create database key-value pair
	kv := keyValue{
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
		if err := <-errCh; err != nil {
			return err
		}
	}

	err = writeRecord(kv, tags)
	if err != nil {
		return
	}
	if del {
		return os.Remove(path)
	}
	return nil
}
