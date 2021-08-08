package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bakape/hydron/v3/assets"
	"github.com/bakape/hydron/v3/common"
	"github.com/bakape/hydron/v3/db"
	"github.com/bakape/hydron/v3/fetch"
	"github.com/bakape/hydron/v3/files"
	"github.com/bakape/hydron/v3/tags"
	"github.com/bakape/hydron/v3/templates"
	"github.com/dimfeld/httptreemux"
	"github.com/gorilla/handlers"
)

var fileHeaders = map[string]string{
	// max-age set to 350 days. Some caches and browsers ignore max-age, if
	// it is a year or greater, so keep it a little below.
	"Cache-Control":   "max-age=30240000, public, immutable",
	"X-Frame-Options": "sameorigin",
	// Fake E-tag, because all files are immutable
	"ETag": "0",
	// To make files downloadable from web browser any page
	"Access-Control-Allow-Origin": "*",
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
	r.GET("/assets/*path", serveAssets)

	// HTML paths
	r.GET("/search", serveSearchHTML)
	r.GET("/image/:id", serveImagePage)
	r.GET("/import", serveImportPage)
	r.GET("/help", serveHelpPage)

	// Image API
	api := r.NewGroup("/api")
	api.GET("/complete_tag/:prefix", completeTagHTTP)
	api.POST("/images/:id/name", setImageNameHTTP)
	api.POST("/import", importPathsHTTP)

	images := api.NewGroup("/images")
	images.GET("/", serveSearch) // Dumps everything
	images.POST("/", importUpload)
	images.GET("/search", serveSearch)

	images.GET("/:id", serveByID)
	images.DELETE("/:id", removeFileHTTP)

	tags := images.NewGroup("/:id/tags")
	tags.PATCH("/", addTagsHTTP)
	tags.POST("/", removeTagsHTTP)
	tags.PATCH("/fetch", fetchTagsHTTP)

	ajax := r.NewGroup("/ajax")
	ajax.GET("/thumbnail/:id", serveThumbnail)

	s := http.Server{
		Addr:    addr,
		Handler: selectiveCompression(r),
	}

	// Stop server on SIGTERM and propagate errors
	errCh := make(chan error)
	go func() {
		term := make(chan os.Signal, 1)
		signal.Notify(term, syscall.SIGTERM)
		<-term
		errCh <- s.Shutdown(context.Background())
	}()
	go func() {
		err := s.ListenAndServe()
		if err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	return <-errCh
}

// Only compress JSON and HTML responses
func selectiveCompression(h http.Handler) http.Handler {
	comp := handlers.CompressHandler(h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch w.Header().Get("Content-Type") {
		case "text/html", "application/json":
			comp.ServeHTTP(w, r)
		default:
			h.ServeHTTP(w, r)
		}
	})
}

func serveAssets(w http.ResponseWriter, r *http.Request) {
	buf, etag, cType, err := assets.Asset("", "/"+extractParam(r, "path"))
	switch err {
	case nil:
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(304)
			return
		}
		setHeaders(w, htmlHeaders)
		h := w.Header()
		h.Set("ETag", etag)
		h.Set("Content-Encoding", "gzip")
		h.Set("Content-Type", cType)
		w.Write(buf)
	case assets.ErrAssetFileNotFound:
		send404(w)
	default:
		send500(w, r, err)
	}
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

// Return data descibing the page the client is requesting
func getRequestPage(r *http.Request) (page common.Page, err error) {
	q := r.URL.Query()
	extractUint := func(key string) uint {
		if s := q.Get(key); s != "" {
			i, _ := strconv.ParseUint(s, 10, 64)
			return uint(i)
		}
		return 0
	}

	page.Page = extractUint("page")
	page.Limit = extractUint("limit")
	page.Order.Type = common.OrderType(extractUint("order"))
	if page.Order.Type > common.Random {
		page.Order.Type = common.None
	}
	page.Order.Reverse = q.Get("reverse") == "on"
	err = tags.ParseFilters(strings.Join(q["q"], " "), &page)
	return
}

// Serve a tag search result as JSON
func serveSearch(w http.ResponseWriter, r *http.Request) {
	err := func() (err error) {
		page, err := getRequestPage(r)
		if err != nil {
			return
		}
		imgs, err := readSearchImages(&page)
		if err != nil {
			return
		}

		// Stream the respone as it is being encoded
		setHeaders(w, jsonHeaders)
		enc := json.NewEncoder(w)
		return enc.Encode(imgs)
	}()
	if err != nil {
		httpError(w, r, err)
	}
}

func readSearchImages(page *common.Page,
) (images []common.CompactImage, err error) {
	images = make([]common.CompactImage, 0, common.PageSize)
	err = db.SearchImages(page, true, func(rec common.CompactImage) error {
		images = append(images, rec)
		return nil
	})
	return
}

func serveSearchHTML(w http.ResponseWriter, r *http.Request) {
	page, err := getRequestPage(r)
	if err != nil {
		httpError(w, r, err)
		return
	}

	// Redirect, if URLs don't match because of extra internal tags input in
	// the search field instead of using the other form input elements
	if r.URL.RawQuery != page.Query() {
		http.Redirect(w, r, page.URL(), 301)
		return
	}

	images, err := readSearchImages(&page)
	if err != nil {
		httpError(w, r, err)
		return
	}

	setHeaders(w, htmlHeaders)
	templates.WriteBrowser(w, page, images)
}

// Serve single image data by ID
func serveByID(w http.ResponseWriter, r *http.Request) {
	img, err := db.GetImage(extractParam(r, "id"))
	if err != nil {
		httpError(w, r, err)
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
	tags, err := db.CompleteTag(extractParam(r, "prefix"))
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

// Set the target file's name
func setImageNameHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		sendError(w, 400, err)
		return
	}

	name := r.Form.Get("name")
	if len(name) > 200 {
		name = name[0:200]
	}

	err = setImageName(extractParam(r, "id"), name)

	switch err {
	case nil:
	case sql.ErrNoRows:
		sendError(w, 400, err)
	default:
		send500(w, r, err)
	}
}

// Fetch tags for a single file
func fetchTagsHTTP(w http.ResponseWriter, r *http.Request) {
	sha1 := extractParam(r, "id")
	pair, err := db.GetImageIDAndMD5(sha1)
	if err != nil {
		send500(w, r, err)
		return
	}

	tags, err := fetch.FetchTags(pair.MD5)
	if err != nil {
		send500(w, r, err)
		return
	}

	err = db.UpdateTags(pair.ID, tags, common.Danbooru)
	switch err {
	case nil:
	case sql.ErrNoRows:
		sendError(w, 400, err)
	default:
		send500(w, r, err)
	}
}

// Serve thumbnail HTML
func serveThumbnail(w http.ResponseWriter, r *http.Request) {
	img, err := db.GetImage(extractParam(r, "id"))
	if err != nil {
		httpError(w, r, err)
		return
	}
	setHeaders(w, htmlHeaders)
	templates.WriteThumbnail(w, img.CompactImage, common.Page{}, false)
}

// Serve expanded view of an image
func serveImagePage(w http.ResponseWriter, r *http.Request) {
	var (
		page common.Page
		img  common.Image
	)
	err := func() (err error) {
		page, err = getRequestPage(r)
		if err != nil {
			return
		}
		img, err = db.GetImage(extractParam(r, "id"))
		return
	}()
	if err != nil {
		httpError(w, r, err)
		return
	}

	setHeaders(w, htmlHeaders)
	templates.WriteImagePage(w, img, page)
}

func serveImportPage(w http.ResponseWriter, r *http.Request) {
	setHeaders(w, htmlHeaders)
	templates.WriteImportPage(w)
}

// Extract import args and pass to importPaths
func importPathsHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		send500(w, r, err)
		return
	}
	paths := parsePaths(r.Form.Get("path"))
	if len(paths) == 0 {
		// Nothing to do, return early
		return
	}
	del, err := strconv.ParseBool(r.Form.Get("del"))
	if err != nil {
		send500(w, r, err)
		return
	}
	fetch, err := strconv.ParseBool(r.Form.Get("fetchTags"))
	if err != nil {
		send500(w, r, err)
		return
	}
	storeName, err := strconv.ParseBool(r.Form.Get("storeName"))
	if err != nil {
		send500(w, r, err)
		return
	}

	err = clientImportPaths(
		w,
		r,
		paths,
		del,
		fetch,
		storeName,
		r.Form.Get("tagStr"),
	)
	if err != nil {
		send500(w, r, err)
		return
	}
}

func serveHelpPage(w http.ResponseWriter, r *http.Request) {
	setHeaders(w, htmlHeaders)
	templates.WriteHelpPage(w)
}
