package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dimfeld/httptreemux"
)

var (
	// Set of headers for serving files
	imageHeaders = map[string]string{
		// max-age set to 350 days. Some caches and browsers ignore max-age, if
		// it is a year or greater, so keep it a little below.
		"Cache-Control":   "max-age=30240000, public, immutable",
		"X-Frame-Options": "sameorigin",
		// Fake E-tag, because all images are immutable
		"ETag": "0",
	}
)

func startServer(addr string) error {
	stderr.Println("listening on " + addr)

	r := httptreemux.New()
	r.PanicHandler = func(
		w http.ResponseWriter,
		_ *http.Request,
		err interface{},
	) {
		sendError(w, 500, fmt.Errorf("%s", err))
	}

	r.GET("/files/*path", serveFiles)

	return http.ListenAndServe(addr, r)
}

// Text-only 404 response
func send404(w http.ResponseWriter) {
	http.Error(w, "404 not found", 404)
}

// Text-only error response
func sendError(w http.ResponseWriter, code int, err error) {
	http.Error(w, fmt.Sprintf("%d %s", code, err), code)
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

	head := w.Header()
	for key, val := range imageHeaders {
		head.Set(key, val)
	}

	http.ServeContent(w, r, p["path"], time.Time{}, file)
}
