package core

import (
	"bytes"
	"sort"
	"strings"
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

// Convert any externally-input tags to the internal format
func normalizeTag(t []byte) []byte {
	for i, b := range t {
		switch b {
		// Replace spaces and NULL with underscores
		case ' ', '\x00':
			t[i] = '_'
		// Simplifies JSON encoding. Tags should not contain quotation marks
		// either way.
		case '"':
			t[i] = '\''
		default:
			// Lowercase letters
			if 'A' <= b && b <= 'Z' {
				t[i] += 'a' - 'A'
			}
		}
	}

	if len(t) > 7 && bytes.HasPrefix(t, []byte("rating:")) {
		// Trim rating tags. Example: `rating:safe` -> `rating:s`
		t = t[:8]
	} else {
		// All other prefixes are stripped to facilitate faster tag searching.
		i := bytes.LastIndexByte(t, ':')
		if i != -1 && !bytes.HasPrefix(t, []byte("system:")) {
			t = t[i+1:]
		}
	}

	return t
}

// Merge two sets of tags
func mergeTagSets(a, b [][]byte) [][]byte {
	set := make(map[string]struct{}, len(a)+len(b))
	for _, arr := range [...][][]byte{a, b} {
		for _, tag := range arr {
			set[string(tag)] = struct{}{}
		}
	}

	return setToTags(set)
}

// Convert an array of tags to a set
func tagsToSet(tags [][]byte) map[string]struct{} {
	set := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		set[string(tag)] = struct{}{}
	}
	return set
}

// Convert a set of tags to an array
func setToTags(set map[string]struct{}) [][]byte {
	tags := make([][]byte, 0, len(set))
	for tag := range set {
		if tag != "" {
			tags = append(tags, []byte(tag))
		}
	}
	return tags
}

// Subtact tags in set b from set a
func subtractTags(a, b [][]byte) [][]byte {
	set := tagsToSet(a)
	for _, tag := range b {
		delete(set, string(tag))
	}
	return setToTags(set)
}

// Return up to 10 random suggestions for tags by contained string
func CompleteTag(s string) []string {
	s = string(normalizeTag([]byte(s)))
	var (
		i    = 0
		tags = make([]string, 0, 10)
	)

	tagMu.RLock()
	for tag := range tagIndex {
		if strings.Contains(tag, s) {
			tags = append(tags, tag)
			i++
		}
		if i == 9 {
			break
		}
	}
	tagMu.RUnlock()

	// Also try to match against system tags
	for _, t := range systemTagHeaders {
		if strings.Contains(t, s) {
			tags = append(tags, "system:"+t)
		}
	}

	sort.Strings(tags)
	return tags
}

// Split a sep-delimited list of tags and normalize each
func SplitTagString(s string, sep byte) [][]byte {
	split := bytes.Split([]byte(s), []byte{sep})
	tags := make([][]byte, 0, len(split))
	for _, tag := range split {
		tag = bytes.TrimSpace(tag)
		if len(tag) == 0 {
			continue
		}
		tags = append(tags, normalizeTag(tag))
	}
	if len(tags) == 0 {
		return nil
	}
	return tags
}
