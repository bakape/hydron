package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
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
	src := make(chan string)
	res := make(chan error)
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
				res <- importFile(f)
			}
		}()
	}

	total := len(files)
	done := 0
	defer stderr.Print("\n")
	for range files {
		switch err := <-res; err {
		case errUnsupportedFile:
			total--
		case nil:
			done++
		default:
			stderr.Print("\n")
			return err
		}
		fmt.Fprintf(
			os.Stderr,
			"\rimporting and thumbnailing: %d / %d - %.2f%%",
			done, total,
			float32(done)/float32(total),
		)
	}

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
	id := hex.EncodeToString(hash[:])

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
		// Silence failed thumbnailing. Perhaps we need better handling?
		err = errUnsupportedFile
		return
	}

	err = ioutil.WriteFile(sourcePath(id, typ), buf, 0600)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(thumbPath(id, isPNG), thumb, 0600)
	if err != nil {
		return
	}

	// TODO: DB stuff

	return
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
