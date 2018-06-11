package main

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/dimfeld/httptreemux"
)

func sendError(w http.ResponseWriter, code int, err error) {
	http.Error(w, fmt.Sprintf("%d %s", code, err), code)
}

func send404(w http.ResponseWriter) {
	http.Error(w, "404 not found", 404)
}

func send500(w http.ResponseWriter, r *http.Request, err interface{}) {
	http.Error(w, fmt.Sprintf("500 %s", err), 500)
	stderr.Printf("server: %s: %s\n%s", r.RemoteAddr, err, debug.Stack())
}

func setHeaders(w http.ResponseWriter, headers map[string]string) {
	head := w.Header()
	for key, val := range headers {
		head.Set(key, val)
	}
}

func setJSONHeaders(w http.ResponseWriter) {
	setHeaders(w, map[string]string{
		"Cache-Control": "max-age=0, must-revalidate",
		"Expires":       "Fri, 01 Jan 1990 00:00:00 GMT",
		"Content-Type":  "application/json",
	})
}

// Adapter for http.HandlerFunc -> httptreemux.HandlerFunc
func wrapHandler(fn http.HandlerFunc) httptreemux.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		fn(w, r)
	}
}
