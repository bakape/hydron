package main

import "C"
import "github.com/mailru/easyjson/jwriter"

// Encodes JSON to a C string
type jsonEncoder struct {
	w jwriter.Writer
}
