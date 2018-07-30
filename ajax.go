package main

import (
	"net/http"
	"strings"

	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/templates"
)

// Serve thumbnail HTML
func serveThumbnail(w http.ResponseWriter, r *http.Request) {
	img, err := db.GetImage(extractParam(r, "id"))
	passQueryError(w, r, err, func() {
		serveHTML(w, r, templates.Thumbnail(img.CompactImage, false))
	})
}

// Serve expanded view of an image
func serveImageView(w http.ResponseWriter, r *http.Request) {
	img, err := db.GetImage(extractParam(r, "id"))
	passQueryError(w, r, err, func() {
		serveHTML(w, r,
			templates.ImageView(img, strings.Join(r.URL.Query()["q"], "")))
	})
}
