package main

import (
	"fmt"
	"net/http"

	"github.com/dimfeld/httptreemux"
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
