//go:generate qtc

package templates

import (
	"bytes"
	"net/url"
	"sort"

	"github.com/bakape/hydron/common"
)

type tagSorter []common.Tag

func (t tagSorter) Len() int {
	return len(t)
}

func (t tagSorter) Less(i, j int) bool {
	return t[i].Tag < t[j].Tag
}

func (t tagSorter) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// Organize and dedup tags based on source and type
func organizeTags(src []common.Tag) map[common.TagType][]common.Tag {
	org := make(map[common.TagType]map[string]common.Tag)

	// Most numerous, so larger preallocation is desirable
	org[common.Undefined] = make(map[string]common.Tag, len(src))

	for _, t := range src {
		ofType := org[t.Type]
		if ofType == nil {
			ofType = make(map[string]common.Tag)
			org[t.Type] = ofType
		}
		existing, ok := ofType[t.Tag]
		// Let higher priority tags override lower priority tags
		if !ok || t.Source < existing.Source {
			ofType[t.Tag] = t
		}
	}

	out := make(map[common.TagType][]common.Tag, len(org))
	for typ, tags := range org {
		sorted := make(tagSorter, 0, len(tags))
		for _, t := range tags {
			sorted = append(sorted, t)
		}
		sort.Sort(sorted)
		out[typ] = []common.Tag(sorted)
	}
	return out
}

func stringifyTagType(t common.TagType) string {
	switch t {
	case common.Author:
		return "author"
	case common.Character:
		return "character"
	case common.Series:
		return "series"
	case common.Rating:
		return "rating"
	default:
		return "undefined"
	}
}

// Add a tag to search query and output query string
func addToQuery(query string, tag common.Tag, substract bool) string {
	buf := bytes.NewBufferString(query)
	buf.WriteByte(' ')
	if substract {
		buf.WriteByte('-')
	}
	if tag.Type != common.Undefined {
		buf.WriteString(stringifyTagType(tag.Type))
		buf.WriteByte(':')
	}
	buf.WriteString(tag.Tag)
	return url.Values{
		"q": []string{buf.String()},
	}.Encode()
}
