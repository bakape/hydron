package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dimfeld/httptreemux"
)

var jsonHeaders = map[string]string{
	"Cache-Control": "max-age=0, must-revalidate",
	"Expires":       "Fri, 01 Jan 1990 00:00:00 GMT",
	"Content-Type":  "application/json",
}

var htmlHeaders = map[string]string{
	"Cache-Control": "max-age=0, must-revalidate",
	"Expires":       "Fri, 01 Jan 1990 00:00:00 GMT",
	"Content-Type":  "text/html",
}

// Error with an attached HTTP status code
type StatusError interface {
	error
	Status() int
}

// Handle error and devulge code from error type or value
func httpError(w http.ResponseWriter, r *http.Request, err error) {
	code := 500
	switch err.(type) {
	case StatusError:
		code = err.(StatusError).Status()
	case *strconv.NumError:
		code = 400
	default:
		if err == sql.ErrNoRows {
			code = 404
		}
	}

	http.Error(w, fmt.Sprintf("%d %s", code, err), code)
	if code >= 500 && code < 600 {
		stderr.Printf("server: %s: %s", r.RemoteAddr, err)
	}
}

func sendError(w http.ResponseWriter, code int, err error) {
	http.Error(w, fmt.Sprintf("%d %s", code, err), code)
}

func send404(w http.ResponseWriter) {
	http.Error(w, "404 not found", 404)
}

func send400(w http.ResponseWriter, err error) {
	sendError(w, 400, err)
}

func send500(w http.ResponseWriter, r *http.Request, err error) {
	sendError(w, 500, err)
	stderr.Printf("server: %s: %s", r.RemoteAddr, err)
}

// Extract URL paramater from request context
func extractParam(r *http.Request, id string) string {
	return httptreemux.ContextParams(r.Context())[id]
}

func setHeaders(w http.ResponseWriter, headers map[string]string) {
	head := w.Header()
	for key, val := range headers {
		head.Set(key, val)
	}
}

func serveJSON(w http.ResponseWriter, r *http.Request, data interface{}) {
	buf, err := json.Marshal(data)
	if err != nil {
		send500(w, r, err)
		return
	}
	serveData(w, r, jsonHeaders, buf)
}

func serveData(w http.ResponseWriter, r *http.Request,
	headers map[string]string, buf []byte,
) {
	h := md5.Sum(buf)
	etag := base64.RawStdEncoding.EncodeToString(h[:])
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(304)
		return
	}

	setHeaders(w, headers)
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.Write(buf)
}
