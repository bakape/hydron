package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dimfeld/httptreemux"
)

func startServer(addr string) error {
	stderr.Println("listening on " + addr)

	r := httptreemux.New()
	r.NotFoundHandler = func(w http.ResponseWriter, _ *http.Request) {
		send404(w)
	}
	r.PanicHandler = func(
		w http.ResponseWriter,
		r *http.Request,
		err interface{},
	) {
		send500(w, r, errors.New(fmt.Sprint(err)))
	}

	// Assets
	r.GET("/files/*path", serveFiles)

	// JSON
	r.GET("/search/:tags", serveSearch)
	r.GET("/search/", func(
		w http.ResponseWriter,
		r *http.Request,
		_ map[string]string,
	) {
		serveSearch(w, r, nil)
	})

	// Commands
	r.GET("/fetch_tags", wrapHandler(fetchAllTagsHTTP))

	// Uploads
	r.POST("/import", wrapHandler(importUpload))

	return http.ListenAndServe(addr, r)
}

// More performant handler for serving file assets. These are immutable, so we
// can also set separate caching policies for them.
func serveFiles(w http.ResponseWriter, r *http.Request, p map[string]string) {
	if r.Header.Get("If-None-Match") == "0" {
		w.WriteHeader(304)
		return
	}

	name := p["path"]
	if len(name) < 40 {
		send404(w)
		return
	}
	path := filepath.Clean(filepath.Join(imageRoot, name[:2], name))
	file, err := os.Open(path)
	if err != nil {
		send404(w)
		return
	}
	defer file.Close()

	setHeaders(w, map[string]string{
		// max-age set to 350 days. Some caches and browsers ignore max-age, if
		// it is a year or greater, so keep it a little below.
		"Cache-Control":   "max-age=30240000, public, immutable",
		"X-Frame-Options": "sameorigin",
		// Fake E-tag, because all files are immutable
		"ETag": "0",
	})

	http.ServeContent(w, r, p["path"], time.Time{}, file)
}

// Serve a tag search result as JSON
func serveSearch(w http.ResponseWriter, r *http.Request, p map[string]string) {
	if p == nil {
		serveAllFileJSON(w, r)
		return
	}

	matched, err := searchByTags(splitTagString(p["tags"], ','))
	if err != nil {
		send500(w, r, err)
		return
	}

	tx, err := db.Begin(false)
	if err != nil {
		send500(w, r, err)
		return
	}

	jrs := newJSONRecordStreamer(w, r)
	defer jrs.close()
	buc := tx.Bucket([]byte("images"))
	for _, id := range matched {
		jrs.writeKeyValue(keyValue{
			sha1:   id,
			record: record(buc.Get(id[:])),
		})
	}
	jrs.err = tx.Rollback()
}

// Serve all file data as JSON
func serveAllFileJSON(w http.ResponseWriter, r *http.Request) {
	jrs := newJSONRecordStreamer(w, r)
	defer jrs.close()
	jrs.err = iterateRecords(func(k []byte, rec record) {
		jrs.writeKeyValue(keyValue{
			sha1:   extractKey(k),
			record: rec,
		})
	})
}

// Download and import a file from the client
func importUpload(w http.ResponseWriter, r *http.Request) {
	f, _, err := r.FormFile("file")
	if err != nil {
		send500(w, r, err)
		return
	}
	defer f.Close()

	// Assign tags to file
	var tags [][]byte
	if s := r.FormValue("tags"); s != "" {
		tags = splitTagString(s, ',')
	}

	kv, err := importFile(f, tags)
	switch err {
	case nil:
	case errImported:
		sendError(w, 409, err)
		return
	case errUnsupportedFile:
		sendError(w, 400, err)
		return
	default:
		send500(w, r, err)
		return
	}

	// Fetch tags from boorus
	if r.FormValue("fetch_tags") != "true" && canFetchTags(kv.record) {
		err := fetchSingleFileTags(kv)
		if err != nil {
			send500(w, r, err)
		}
	}
}

// Attempt to fetch tags for all files that have not yet had their tags synced
// with gelbooru.com
func fetchAllTagsHTTP(w http.ResponseWriter, r *http.Request) {
	err := fetchAllTags()
	if err != nil {
		send500(w, r, err)
	}
}
