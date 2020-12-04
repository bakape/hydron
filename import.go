package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/files"
	imp "github.com/bakape/hydron/import"
)

/*
Import file/directory paths
paths: list for file and/or directory paths
del: delete files after import
fetchTags: also fetch tags from danbooru
tagStr: add string of tags to all imported files
*/
func importPaths(paths []string, del, fetchTags bool, tagStr string) error {
	paths, err := files.Traverse(paths)
	if err != nil {
		return err
	}

	// Buffer all paths into a channel
	passPaths := make(chan string, len(paths))
	for _, p := range paths {
		passPaths <- p
	}

	// Process files in parallel
	ch := make(chan Result)
	n := runtime.NumCPU() + 1
	for i := 0; i < n; i++ {
		go func() {
			for p := range passPaths {
				res := new(Result)
				_, res.err = importPath(p, del, fetchTags, tagStr)
				ch <- *res
			}
		}()
	}

	// Aggregate and log results
	p := progressLogger{
		header: "importing, thumbnailing",
		total:  len(paths),
	}
	if fetchTags {
		p.header += ", fetching tags"
	}
	for i := 0; i < len(paths); i++ {
		store := new(Result)
		*store = <-ch
		err = store.err
		if err != nil {
			p.Err(err)
		} else {
			p.Done()
		}
	}
	close(passPaths) // Stop worker goroutines
	p.Close()

	return nil
}

func importPath(p string, del, fetchTags bool, tagStr string) (img common.Image, err error) {
	f, err := os.Open(p)
	if err != nil {
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return
	}

	name := info.Name()
	img, err = imp.ImportFile(
		f,
		int(info.Size()),
		strings.TrimSuffix(name, filepath.Ext(name)),
		tagStr,
		fetchTags,
	)

	switch err {
	case nil:
	case imp.ErrImported:
		err = nil
	default:
		err = fmt.Errorf("%s: %s", p, err)
		return
	}

	if del {
		// Close file before removing
		f.Close()
		f = nil
		err = os.Remove(p)
	}
	return
}

// Download and import a file from the client
func importUpload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(0)
	if err != nil {
		sendError(w, 400, err)
		return
	}

	f, info, err := r.FormFile("file")
	if err != nil {
		sendError(w, 400, err)
		return
	}
	defer f.Close()

	size, err := strconv.ParseUint(r.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		sendError(w, 400, err)
		return
	}

	img, err := imp.ImportFile(
		f,
		int(size),
		strings.TrimSuffix(info.Filename, filepath.Ext(info.Filename)),
		r.Form.Get("tags"),
		r.Form.Get("fetch_tags") == "true",
	)

	switch err {
	case nil:
	case imp.ErrImported:
		// Still return the JSON, if already imported
		img, err = db.GetImage(img.SHA1)
		if err != nil {
			send500(w, r, err)
			return
		}
	case imp.ErrUnsupportedFile:
		sendError(w, 415, err)
		return
	default:
		send500(w, r, err)
		return
	}

	serveJSON(w, r, img)
}

// Copy of importPaths that sends progress to client
// Temporary name, think of a better one
func clientImportPaths(w http.ResponseWriter, r *http.Request, paths []string, del, fetchTags bool, tagStr string) error {
	paths, err := files.Traverse(paths)
	if err != nil {
		return err
	}

	// Buffer all paths into a channel
	passPaths := make(chan string, len(paths))
	for _, p := range paths {
		passPaths <- p
	}

	// Process files in parallel
	ch := make(chan Result)
	n := runtime.NumCPU() + 1
	for i := 0; i < n; i++ {
		go func() {
			for p := range passPaths {
				res := new(Result)
				res.img, res.err = importPath(p, del, fetchTags, tagStr)
				ch <- *res
			}
		}()
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Fatal("Failed to make HTTP Flusher.")
	}
	total := len(paths)

	for i := 0; i < total; i++ {
		store := new(Result)
		*store = <-ch
		err = store.err

		switch err {
		case nil:
		case imp.ErrImported:
		case imp.ErrUnsupportedFile:
			sendError(w, 415, err)
			return err
		default:
			send500(w, r, err)
			return err
		}

		jsonChunk, err := json.Marshal(Chunk{store.img.SHA1, i + 1, total})
		if err != nil {
			send500(w, r, err)
			return err
		}
		// Need to separate chunks, hyphen never appears in chunk content
		// so is safe to use
		fmt.Fprintf(w, "%s-", jsonChunk)
		flusher.Flush()
	}
	close(passPaths) // Stop worker goroutines

	return nil
}
