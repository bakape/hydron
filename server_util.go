package main

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/mailru/easyjson/jwriter"
)

// Helper for continuously streaming JSON file data to the client as it is being
// retrieved and encoded
type jsonRecordStreamer struct {
	minimal bool
	jwriter.Writer
	w   http.ResponseWriter
	r   *http.Request
	err error
}

func newJSONRecordStreamer(
	w http.ResponseWriter,
	r *http.Request,
) jsonRecordStreamer {
	setJSONHeaders(w)
	w.Write([]byte{'['})
	return jsonRecordStreamer{
		minimal: r.URL.Query().Get("minimal") == "true",
		w:       w,
		r:       r,
	}
}

func (j *jsonRecordStreamer) close() {
	if j.err != nil {
		send500(j.w, j.r, j.err)
		return
	}
	j.w.Write([]byte{']'})
}

func (j *jsonRecordStreamer) writeKeyValue(kv keyValue) {
	kv.toJSON(&j.Writer, j.minimal)
	j.DumpTo(j.w)
}

func send400(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("400 %s", err), 400)
}

func send403(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("403 %s", err), 403)
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
