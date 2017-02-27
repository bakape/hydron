package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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
func importPaths(paths []string, del bool) (err error) {
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
	for i := 0; i < n; i++ {
		go func() {
			for f := range src {
				res <- response{
					path: f,
					err:  importFile(f, del),
				}
			}
		}()
	}

	total := len(files)
	done := 0
	for range files {
		switch r := <-res; r.err {
		case errUnsupportedFile:
			total--
		case nil:
			done++
		default:
			total--
			stderr.Printf("\n%s: %s\n", r.err, r.path)
		}
		fmt.Fprintf(
			os.Stderr,
			"\rimporting and thumbnailing: %d / %d - %.2f%%",
			done, total,
			float32(done)/float32(total)*100,
		)
	}
	stderr.Print("\n")

	return nil
}

func importFile(path string, del bool) (err error) {
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
		errCh <- ioutil.WriteFile(path, buf, 0600)
	}()
	go func() {
		path := filepath.FromSlash(thumbPath(id, thumb.IsPNG))
		errCh <- ioutil.WriteFile(path, thumb.Data, 0600)
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

	err = writeRecord(kv)
	if err != nil {
		return
	}
	if del {
		return os.Remove(path)
	}
	return nil
}

func detectFileType(r io.ReadSeeker) (fileType, error) {
	buf := make([]byte, 512)
	read, err := r.Read(buf)
	if err != nil {
		return 0, err
	}
	t, ok := mimeTypes[http.DetectContentType(buf[:read])]
	if !ok {
		return 0, errUnsupportedFile
	}
	return t, nil
}
