package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/files"
	"github.com/bakape/hydron/import"
	"github.com/bakape/hydron/tags"
	"github.com/bakape/hydron/templates"
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

	r.GET("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/search", 301)
	})

	// Assets
	r.GET("/files/:file", func(w http.ResponseWriter, r *http.Request) {
		serveFiles(w, r, files.ImageRoot)
	})
	r.GET("/thumbs/:file", func(w http.ResponseWriter, r *http.Request) {
		serveFiles(w, r, files.ThumbRoot)
	})
	r.GET("/assets/*path", func(w http.ResponseWriter, r *http.Request) {
		serveFile(w, r,
			filepath.Clean(filepath.Join("www", extractParam(r, "path"))))
	})

	// HTML paths
	r.GET("/search", serveSearchHTML)

	// Image API
	api := r.NewGroup("/api")
	api.GET("/complete_tag/:prefix", completeTagHTTP)

	images := api.NewGroup("/images")
	images.GET("/", serveSearch) // Dumps everything
	images.POST("/", importUpload)
	images.GET("/search", serveSearch)

	images.GET("/:id", serveByID)
	images.DELETE("/:id", removeFileHTTP)

	tags := images.NewGroup("/:id/tags")
	tags.PATCH("/", addTagsHTTP)
	tags.DELETE("/", removeTagsHTTP)

	return http.ListenAndServe(addr, selectiveCompression(r))
}

// Only compress JSON and HTML responses
func selectiveCompression(h http.Handler) http.Handler {
	comp := handlers.CompressHandler(h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch w.Header().Get("Content-Type") {
		case "text/html", "application/json", "application/javascript",
			"text/css":
			comp.ServeHTTP(w, r)
		default:
			h.ServeHTTP(w, r)
		}
	})
}

func serveFile(w http.ResponseWriter, r *http.Request, path string) {
	file, err := os.Open(path)
	if err != nil {
		send404(w)
		return
	}
	defer file.Close()

	stats, err := file.Stat()
	if err != nil {
		send500(w, r, err)
		return
	}
	if stats.IsDir() {
		send404(w)
		return
	}
	modTime := stats.ModTime()
	etag := strconv.FormatInt(modTime.Unix(), 10)

	head := w.Header()
	head.Set("Cache-Control", "no-cache")
	head.Set("ETag", etag)
	http.ServeContent(w, r, path, modTime, file)
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

// If ok = false, caller should return too
func processSearch(w http.ResponseWriter, r *http.Request,
	fn func(common.CompactImage) error,
) (params string, page, totalPages int, ok bool) {
	if s := r.URL.Query().Get("page"); s != "" {
		page, _ = strconv.Atoi(s)
		if page < 0 {
			page = 0
		}
	}
	params = strings.Join(r.URL.Query()["q"], " ")

	totalPages, err := db.SearchImages(params, page, fn)
	switch err.(type) {
	case nil:
		ok = true
		return
	case tags.SyntaxError:
		sendError(w, 400, err)
		return
	default:
		send500(w, r, err)
		return
	}
}

// Serve a tag search result as JSON
func serveSearch(w http.ResponseWriter, r *http.Request) {
	var jw jwriter.Writer
	jw.RawByte('[')
	first := true
	_, _, _, ok := processSearch(w, r, func(rec common.CompactImage) error {
		if first {
			first = false
		} else {
			jw.RawByte(',')
		}
		rec.MarshalEasyJSON(&jw)
		return nil
	})
	if !ok {
		return
	}
	jw.RawByte(']')

	buf, err := jw.BuildBytes()
	if err != nil {
		send500(w, r, err)
		return
	}
	serveData(w, r, jsonHeaders, buf)
}

func serveSearchHTML(w http.ResponseWriter, r *http.Request) {
	images := make([]common.CompactImage, 0, 1<<10)
	params, page, pages, ok := processSearch(w, r,
		func(rec common.CompactImage) error {
			images = append(images, rec)
			return nil
		})
	if ok {
		serveHTML(w, r, templates.Browser(params, page, pages, images))
	}
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
	f, _, err := r.FormFile("file")
	if err != nil {
		sendError(w, 400, err)
		return
	}
	defer f.Close()

	img, err := imp.ImportFile(f, r.Form.Get("tags"),
		r.Form.Get("fetch_tags") == "true")
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

// Remove files from the database by ID
func removeFileHTTP(w http.ResponseWriter, r *http.Request) {
	err := db.RemoveImage(extractParam(r, "id"))
	if err != nil {
		send500(w, r, err)
	}
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

	serveData(w, r, jsonHeaders, b)
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
