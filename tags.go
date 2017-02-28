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
			file[j] = list[i*20+j]
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

// Convert any externally-input tags to the internal format
func normalizeTag(t []byte) []byte {
	for i, b := range t {
		switch b {
		// Replace spaces and NULL with underscores
		case ' ', '\x00':
			t[i] = '_'
		default:
			// Lowercase letters
			if 'A' <= b && b <= 'Z' {
				t[i] += 'a' - 'A'
			}
		}
	}

	// Trim rating tags. Example: `rating:safe` -> `rating:s`
	if len(t) > 7 && bytes.HasPrefix(t, []byte("rating:")) {
		t = t[:8]
	}

	return t
}

// Merge two sets of tags
func mergeTagSets(a, b [][]byte) [][]byte {
	tags := make(map[string]struct{}, len(a)+len(b))
	for _, set := range [...][][]byte{a, b} {
		for _, tag := range set {
			tags[string(tag)] = struct{}{}
		}
	}

	merged := make([][]byte, 0, len(tags))
	for tag := range tags {
		if len(tag) != 0 {
			merged = append(merged, []byte(tag))
		}
	}
	return merged
}

// Find files that match all tags
func searchByTags(tags [][]byte) [][20]byte {
	if len(tags) == 0 {
		return nil
	}
	tagMu.RLock()
	defer tagMu.RUnlock()

	first := tagIndex[string(tags[0])]
	if first == nil {
		return nil
	}

	// Copy map. Original must not be modified.
	matched := make(map[[20]byte]bool, len(first))
	for f := range first {
		matched[f] = true
	}

	// Delete non-intersecting matches
	for i := 1; i < len(tags); i++ {
		files := tagIndex[string(tags[i])]
		if files == nil {
			return nil
		}
		for f := range matched {
			if !files[f] {
				delete(matched, f)
			}
		}
	}

	arr := make([][20]byte, 0, len(matched))
	for f := range matched {
		arr = append(arr, f)
	}
	return arr
}
