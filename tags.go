package main

import (
	"bytes"
	"sync"
)

// Tags indexes are kept in memory and dumped into the database on update
var (
	tagIndex = make(map[string]map[[20]byte]bool, 512)
	tagMu    sync.RWMutex
)

// Add a tag->file relation to the index
func indexTag(tag []byte, file [20]byte) {
	tagMu.Lock()
	defer tagMu.Unlock()
	indexTagNoMu(tag, file)
}

// Add a tag->file relation to the index without locking the mutex.
// Used during startup or by manually locking the mutex.
func indexTagNoMu(tag []byte, file [20]byte) {
	key := string(tag)
	t, ok := tagIndex[key]
	if !ok {
		t = make(map[[20]byte]bool, 4)
		tagIndex[key] = t
	}
	t[file] = true
}

// Remove a file->tags relation from the index
func unindexFile(file [20]byte, tags [][]byte) {
	tagMu.Lock()
	defer tagMu.Unlock()
	unindexFileNoMu(file, tags)
}

// Like unindexFile, but does not lock the mutex.
func unindexFileNoMu(file [20]byte, tags [][]byte) {
	for _, tag := range tags {
		key := string(tag)
		t := tagIndex[key]
		delete(t, file)
		if len(t) == 0 {
			delete(tagIndex, key)
		}
	}
}

// Decode list of tagged files
func decodeTaggedList(list []byte) [][20]byte {
	l := len(list) / 20
	files := make([][20]byte, l)
	for i := 0; i < l; i++ {
		var file [20]byte
		for j := 0; j < 20; j++ {
			file[j] = file[i*20+j]
		}
		files[i] = file
	}
	return files
}

// Encode a list of tagged files for storage in the database
func encodeTaggedList(tagged map[[20]byte]bool) []byte {
	l := len(tagged)
	enc := make([]byte, 0, l*20)
	for file := range tagged {
		enc = append(enc, file[:]...)
	}
	return enc
}

// Split a space delimited list of tags and normalize each
func splitTagString(s string) [][]byte {
	split := bytes.Split([]byte(s), []byte{' '})
	tags := make([][]byte, 0, len(split))
	for _, tag := range split {
		tag = bytes.TrimSpace(tag)
		if len(tag) == 0 {
			continue
		}
		tags = append(tags, normalizeTag(tag))
	}
	return tags
}

func normalizeTag(t []byte) []byte {
	// Replace spaces and NULL with underscores
	for i, b := range t {
		switch b {
		case ' ', '\x00':
			t[i] = '_'
		}
	}

	// Trim rating tags. Example: `rating:safe` -> `rating:s`
	if len(t) > 7 && bytes.HasPrefix(t, []byte("rating:")) {
		t = t[:8]
	}

	return t
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
