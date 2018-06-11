package main

import (
	"compress/gzip"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/files"
	"github.com/dimfeld/httptreemux"
	"github.com/gorilla/handlers"
	"github.com/mailru/easyjson/jwriter"
)

func startServer(addr string) error {
	stderr.Println("listening on " + addr)

	r := httptreemux.NewContextMux()
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
	r.GET("/files/:file", func(w http.ResponseWriter, r *http.Request) {
		serveFiles(w, r, files.ImageRoot)
	})
	r.GET("/thumbs/:file", func(w http.ResponseWriter, r *http.Request) {
		serveFiles(w, r, files.ThumbRoot)
	})

	// Image API

	r.GET("/complete_tag/:prefix", completeTagHTTP)

	images := r.NewContextGroup("/images")
	images.GET("/", serveEverything) // Dumps everything
	images.POST("/", importUpload)

	search := images.NewContextGroup("/search")
	search.GET("/", serveEverything)
	search.GET("/:tags", func(w http.ResponseWriter, r *http.Request) {
		q := html.UnescapeString(extractParam(r, "tags"))
		q = strings.Replace(q, ",", " ", -1)
		serveSearch(w, r, q)
	})

	imgID := images.NewContextGroup("/:id")
	imgID.GET("/", getFilesByIDs)
	imgID.DELETE("/", removeFilesHTTP)

	tags := imgID.NewContextGroup("/tags")
	tags.PATCH("/", addTagsHTTP)
	tags.DELETE("/", removeTagsHTTP)

	return http.ListenAndServe(addr,
		handlers.CompressHandlerLevel(r, gzip.DefaultCompression))
}

// Extract URL paramater from request context
func extractParam(r *http.Request, id string) string {
	return httptreemux.ContextParams(r.Context())[id]
}

// More performant handler for serving file assets. These are immutable, so we
// can also set separate caching policies for them.
func serveFiles(w http.ResponseWriter, r *http.Request, root string) {
	if r.Header.Get("If-None-Match") == "0" {
		w.WriteHeader(304)
		return
	}

	name := extractParam(r, "name")
	if len(name) < 40 {
		send404(w)
		return
	}
	path := filepath.Join(root, name[:2], name)
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

	http.ServeContent(w, r, extractParam(r, "path"), time.Time{}, file)
}

// Dump all images as JSON
func serveEverything(w http.ResponseWriter, r *http.Request) {
	serveSearch(w, r, "")
}

// Serve a tag search result as JSON
func serveSearch(w http.ResponseWriter, r *http.Request, params string) {
	var jw jwriter.Writer
	jw.RawByte('[')
	first := true
	err := db.SearchImages(params, func(rec common.CompactImage) error {
		if first {
			first = false
		} else {
			jw.RawByte(',')
		}
		rec.MarshalEasyJSON(&jw)
		return nil
	})
	if err != nil {
		send500(w, r, err)
		return
	}
	jw.RawByte(']')

	setJSONHeaders(w)
	jw.DumpTo(w)
}

// Retrieve records from the database and serve as JSON
func serveJSONByID(
	w http.ResponseWriter,
	r *http.Request,
	ids map[[20]byte]bool,
) {
	panic("TODO")
	// jrs := newJSONRecordStreamer(w, r)
	// defer jrs.close()
	// jrs.err = core.MapRecords(ids, func(id [20]byte, r core.Record) {
	// 	jrs.writeKeyValue(core.KeyValue{
	// 		SHA1:   id,
	// 		Record: r,
	// 	})
	// })
}

// Download and import a file from the client
func importUpload(w http.ResponseWriter, r *http.Request) {
	panic("TODO")
	// f, _, err := r.FormFile("file")
	// if err != nil {
	// 	send500(w, r, err)
	// 	return
	// }
	// defer f.Close()

	// // Assign tags to file
	// var tags [][]byte
	// if s := r.FormValue("tags"); s != "" {
	// 	tags = tags.SplitTagString(s, ',')
	// }

	// kv, err := import.ImportFile(f, tags)
	// switch err {
	// case nil:
	// case import.ErrImported:
	// 	sendError(w, 409, err)
	// 	return
	// case import.ErrUnsupportedFile:
	// 	sendError(w, 400, err)
	// 	return
	// default:
	// 	send500(w, r, err)
	// 	return
	// }

	// // Fetch tags from boorus
	// if r.FormValue("fetch_tags") != "true" && core.CanFetchTags(kv.Record) {
	// 	err := core.FetchSingleFileTags(kv)
	// 	if err != nil {
	// 		send500(w, r, err)
	// 	}
	// }
}

// Remove files from the database by ID
func removeFilesHTTP(w http.ResponseWriter, r *http.Request) {
	panic("TODO")
	// err := removeFiles(strings.Split(p["ids"], ","))
	// if err != nil {
	// 	_, ok := err.(core.InvalidIDError)
	// 	if ok {
	// 		sendError(w, 400, err)
	// 	} else {
	// 		send500(w, r, err)
	// 	}
	// }
}

// Serve file JSON by ID
func getFilesByIDs(w http.ResponseWriter, r *http.Request) {
	panic("TODO")
	// split := strings.Split(p["ids"], ",")
	// ids := make(map[[20]byte]bool, len(split))
	// for i := range split {
	// 	id, err := core.StringToSHA1(split[i])
	// 	if err != nil {
	// 		sendError(w, 400, err)
	// 		return
	// 	}
	// 	ids[id] = true
	// }

	// serveJSONByID(w, r, ids)
}

// Complete a tag by prefix from an HTTP request
func completeTagHTTP(w http.ResponseWriter, r *http.Request) {
	setJSONHeaders(w)

	tags, err := db.CompleTag(extractParam(r, "prefix"))
	if err != nil {
		send500(w, r, err)
		return
	}

	b := make([]byte, 1, 256)
	b[0] = '['
	for i, t := range tags {
		if i != 0 {
			b = append(b, ',')
		}
		b = strconv.AppendQuote(b, t)
	}
	b = append(b, ']')

	w.Write(b)
}

// Add tags to a specific file
func addTagsHTTP(w http.ResponseWriter, r *http.Request) {
	modTagsHTTP(w, r, addTags)
}

func modTagsHTTP(w http.ResponseWriter, r *http.Request,
	fn func(sha1 string, tagStr string) error,
) {
	err := r.ParseForm()
	if err != nil {
		sendError(w, 400, err)
		return
	}

	err = fn(extractParam(r, "id"), r.Form.Get("tags"))
	switch err {
	case nil:
	case sql.ErrNoRows:
		sendError(w, 400, err)
	default:
		send500(w, r, err)
	}
}

// Remove tags from a specific file
func removeTagsHTTP(w http.ResponseWriter, r *http.Request) {
	modTagsHTTP(w, r, removeTags)
}
