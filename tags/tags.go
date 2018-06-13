package tags

import (
	"strings"

	"github.com/bakape/hydron/common"
)

// Convert any externally-input tags to the internal format
// source: tag source to apply to tag
func Normalize(s string, source common.TagSource) (
	tag common.Tag,
) {
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
		case "rating":
			tag.Type = common.Rating
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
