//go:generate qtc

package templates

import (
	"sort"

	"github.com/bakape/hydron/common"
)

// Human-readable labels fo``r image ordering types
var orderLabels = [...]string{"None", "Size", "Width", "Height", "Duration",
	"Tag count", "Random"}

// Human-readable labels for option types
var optionLabels = [...]string{"Fetch tags", "Add tags", "Remove tags",
	"Set name", "Delete files"}

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
