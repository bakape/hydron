package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"path/filepath"

	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/files"
	imp "github.com/bakape/hydron/import"
)

/*
Import file/directory paths
paths: list for file and/pr directory paths
del: delete files after import
fetchTags: also fetch tags from gelbooru
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
	ch := make(chan error)
	n := runtime.NumCPU() + 1
	for i := 0; i < n; i++ {
		go func() {
			for p := range passPaths {
				ch <- importPath(p, del, fetchTags, tagStr)
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
		err = <-ch
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

func importPath(p string, del, fetchTags bool, tagStr string) (err error) {
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
	_, err = imp.ImportFile(
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
