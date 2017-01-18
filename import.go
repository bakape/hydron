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
	"runtime"
	"time"
)

type fileType uint8

// Supported file types
const (
	JPEG fileType = iota
	PNG
	GIF
	WEBM
	PDF
	_
	MP4
	_
	OGG
)

const fileMode = os.O_WRONLY | os.O_CREATE | os.O_EXCL

var (
	errUnsupportedFile = errors.New("usuported file type")

	// Map of net/http MIME types to the constants used internally
	mimeTypes = map[string]fileType{
		"image/jpeg":      JPEG,
		"image/png":       PNG,
		"image/gif":       GIF,
		"video/webm":      WEBM,
		"application/pdf": PDF,
		"application/ogg": OGG,
		"video/mp4":       MP4,
	}
)

// Recursively import list of files and/or directories
func importPaths(paths []string) (err error) {
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
					err:  importFile(f),
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

func importFile(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	typ, err := detectFileType(f)
	if err != nil {
		return
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		return
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	hash := md5.Sum(buf)

	// Check, if not already in the database
	if isImp, err := isImported(hash); err != nil || isImp {
		return err
	}

	var fn func([]byte) ([]byte, bool, error)
	switch typ {
	case WEBM:
		fn = processWebm
	case MP4, OGG:
		fn = processMediaContainer
	default:
		fn = processImage
	}
	thumb, isPNG, err := fn(buf)
	if err != nil {
		return
	}

	errCh := make(chan error)
	sha1Ch := make(chan [20]byte)
	id := hex.EncodeToString(hash[:])
	go func() {
		errCh <- ioutil.WriteFile(sourcePath(id, typ), buf, 0600)
	}()
	go func() {
		sha1Ch <- sha1.Sum(buf)
	}()
	err = ioutil.WriteFile(thumbPath(id, isPNG), thumb, 0600)
	if err != nil {
		return
	}
	if err = <-errCh; err != nil {
		return err
	}

	// Write to database
	kv := keyValue{
		MD5:    hash,
		record: make(record, 31),
	}
	kv.SetImportTime(uint64(time.Now().Unix()))
	kv.SetType(typ)
	if isPNG {
		kv.SetThumbType(PNG)
	} else {
		kv.SetThumbType(JPEG)
	}
	kv.SetSHA1(<-sha1Ch)
	return writeRecord(kv)
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
