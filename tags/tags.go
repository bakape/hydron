package tags

import (
	"strings"

	"github.com/bakape/hydron/common"
)

// Convert any externally-input tags to the internal format
func Normalize(s string, source common.TagSource) (tag common.Tag) {
	tag.Source = source

	i := strings.IndexByte(s, ':')
	if i != -1 {
		switch s[:i] {
		case "artist", "author":
			tag.Type = common.Author
		case "series", "copyright":
			tag.Type = common.Series
		case "character":
			tag.Type = common.Character
		}
		s = s[strings.LastIndexByte(s, ':')+1:]
	}

	buf := []byte(s)
	for i, b := range buf {
		switch b {
		// Replace spaces and NULL with underscores
		case ' ', '\x00':
			buf[i] = '_'
		// Simplifies JSON encoding. Tags should not contain quotation marks
		// either way.
		case '"':
			buf[i] = '\''
		default:
			// Lowercase letters
			if 'A' <= b && b <= 'Z' {
				buf[i] += 'a' - 'A'
			}
		}
	}
	tag.Tag = string(buf)

	return
}

// // Merge two sets of tags
// func mergeTagSets(a, b [][]byte) [][]byte {
// 	set := make(map[string]struct{}, len(a)+len(b))
// 	for _, arr := range [...][][]byte{a, b} {
// 		for _, tag := range arr {
// 			set[string(tag)] = struct{}{}
// 		}
// 	}

// 	return setToTags(set)
// }

// // Convert an array of tags to a set
// func tagsToSet(tags [][]byte) map[string]struct{} {
// 	set := make(map[string]struct{}, len(tags))
// 	for _, tag := range tags {
// 		set[string(tag)] = struct{}{}
// 	}
// 	return set
// }

// // Convert a set of tags to an array
// func setToTags(set map[string]struct{}) [][]byte {
// 	tags := make([][]byte, 0, len(set))
// 	for tag := range set {
// 		if tag != "" {
// 			tags = append(tags, []byte(tag))
// 		}
// 	}
// 	return tags
// }

// // Subtact tags in set b from set a
// func subtractTags(a, b [][]byte) [][]byte {
// 	set := tagsToSet(a)
// 	for _, tag := range b {
// 		delete(set, string(tag))
// 	}
// 	return setToTags(set)
// }

// Split a sep-delimited list of tags and normalize each
func FromString(s string, source common.TagSource) []common.Tag {
	split := strings.Split(s, " ")
	tags := make([]common.Tag, 0, len(split))
	for _, tag := range split {
		tag = strings.TrimSpace(tag)
		if len(tag) == 0 {
			continue
		}
		norm := Normalize(tag, source)
		if norm.Tag != "" {
			tags = append(tags, norm)
		}
	}
	return tags
}
