package main

import (
	"net/http"

	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/templates"
)

// Serve thumbnail HTML
func serveThumbnail(w http.ResponseWriter, r *http.Request) {
	img, err := db.GetImage(extractParam(r, "id"))
	if err != nil {
		httpError(w, r, err)
		return
	}
	setHeaders(w, htmlHeaders)
	templates.WriteThumbnail(w, img.CompactImage, false)
}

// Serve expanded view of an image
func serveImageView(w http.ResponseWriter, r *http.Request) {
	page, err := getRequestPage(r)
	if err != nil {
		httpError(w, r, err)
		return
	}
	setHeaders(w, htmlHeaders)
	templates.WriteImageView(w, page)
}
