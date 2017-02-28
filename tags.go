package main

import (
	"bytes"
)

func normalizeTag(t *[]byte) {
	// Replace spaces and NULL with underscores
	for i, b := range *t {
		switch b {
		case ' ', '\x00':
			(*t)[i] = '_'
		}
	}

	// Trim rating tags. Example: `rating:safe` -> `rating:s`
	if len(*t) > 7 && bytes.HasPrefix(*t, []byte("rating:")) {
		*t = (*t)[:8]
	}
}

func mergeTagSets(a, b [][]byte) [][]byte {
	tags := make(map[string]struct{}, len(a)+len(b))
	for _, set := range [...][][]byte{a, b} {
		for _, tag := range set {
			tags[string(tag)] = struct{}{}
		}
	}

	merged := make([][]byte, 0, len(tags))
	for tag := range tags {
		merged = append(merged, []byte(tag))
	}
	return merged
}
