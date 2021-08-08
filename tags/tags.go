package tags

import (
	"strings"

	"github.com/bakape/hydron/v3/common"
)

// Convert any externally-input tags to the internal format
// source: tag source to apply to tag
func Normalize(s string, source common.TagSource) common.Tag {
	return common.Tag{
		Source: source,
		TagBase: common.TagBase{
			Type: detectTagType(&s),
			Tag:  normalizeString(s),
		},
	}
}

func detectTagType(s *string) (typ common.TagType) {
	i := strings.IndexByte(*s, ':')
	if i != -1 {
		switch (*s)[:i] {
		case "undefined":
			typ = common.Undefined
		case "artist", "author":
			typ = common.Author
		case "series", "copyright":
			typ = common.Series
		case "character":
			typ = common.Character
		case "rating":
			typ = common.Rating
		case "meta":
			typ = common.Meta
		default:
			return
		}
		*s = (*s)[i+1:]
	}
	return
}

func normalizeString(s string) string {
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
	return string(buf)
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
