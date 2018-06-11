package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
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

func serveJSON(w http.ResponseWriter, r *http.Request, data interface{}) {
	buf, err := json.Marshal(data)
	if err != nil {
		send500(w, r, err)
		return
	}
	setJSONHeaders(w)
	w.Write(buf)
}
