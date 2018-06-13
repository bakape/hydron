package main

import (
	"database/sql"
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/bakape/hydron/tags"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/files"
	"github.com/dimfeld/httptreemux"
	"github.com/gorilla/handlers"
	"github.com/mailru/easyjson/jwriter"
)

var fileHeaders = map[string]string{
	// max-age set to 350 days. Some caches and browsers ignore max-age, if
	// it is a year or greater, so keep it a little below.
	"Cache-Control":   "max-age=30240000, public, immutable",
	"X-Frame-Options": "sameorigin",
	// Fake E-tag, because all files are immutable
	"ETag": "0",
}

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
		http.Error(w, fmt.Sprintf("500 %s", err), 500)
		stderr.Printf("server: %s: %s\n%s", r.RemoteAddr, err, debug.Stack())
	}

	// Assets
	r.GET("/files/:file", func(w http.ResponseWriter, r *http.Request) {
		serveFiles(w, r, files.ImageRoot)
	})
	r.GET("/thumbs/:file", func(w http.ResponseWriter, r *http.Request) {
		serveFiles(w, r, files.ThumbRoot)
	})

	// Image API
	api := r.NewGroup("/api")
	api.GET("/complete_tag/:prefix", completeTagHTTP)

	images := api.NewGroup("/images")
	images.GET("/", serveEverything) // Dumps everything
	images.POST("/", importUpload)

	search := images.NewGroup("/search")
	search.GET("/", serveEverything)
	search.GET("/:tags", func(w http.ResponseWriter, r *http.Request) {
		q := html.UnescapeString(extractParam(r, "tags"))
		q = strings.Replace(q, ",", " ", -1)
		serveSearch(w, r, q)
	})

	images.GET("/:id", serveByID)
	images.DELETE("/:id", removeFilesHTTP)

	tags := images.NewGroup("/:id/tags")
	tags.PATCH("/", addTagsHTTP)
	tags.DELETE("/", removeTagsHTTP)

	return http.ListenAndServe(addr, compressAPI(r))
}

// Only compress API responses
func compressAPI(h http.Handler) http.Handler {
	comp := handlers.CompressHandler(h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			comp.ServeHTTP(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
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

	name := extractParam(r, "file")
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

	setHeaders(w, fileHeaders)

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
	switch err.(type) {
	case nil:
	case tags.SyntaxError:
		sendError(w, 400, err)
		return
	default:
		send500(w, r, err)
		return
	}
	jw.RawByte(']')

	buf, err := jw.BuildBytes()
	if err != nil {
		send500(w, r, err)
		return
	}
	serveJSONBuf(w, r, buf)
}

// Serve single image data by ID
func serveByID(w http.ResponseWriter, r *http.Request) {
	img, err := db.GetImage(extractParam(r, "id"))
	switch err {
	case nil:
		serveJSON(w, r, img)
	case sql.ErrNoRows:
		send404(w)
	default:
		send500(w, r, err)
	}
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

// Complete a tag by prefix from an HTTP request
func completeTagHTTP(w http.ResponseWriter, r *http.Request) {
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

	serveJSONBuf(w, r, b)
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
