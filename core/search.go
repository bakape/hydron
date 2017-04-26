package core

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Find files that match all tags
func SearchByTags(tags [][]byte) (matched map[[20]byte]bool, err error) {
	if len(tags) == 0 {
		return
	}

	// Separate system tags from regular tags
	regular := make([][]byte, 0, len(tags))
	system := make([][]byte, 0, len(tags))
	for _, t := range tags {
		if isSystemTag(t) {
			system = append(system, t)
		} else {
			regular = append(regular, t)
		}
	}

	tagMu.RLock()
	defer tagMu.RUnlock()

	if len(regular) != 0 {
		first := tagIndex[string(regular[0])]
		if first == nil {
			return
		}

		// Copy map. Original must not be modified.
		matched = make(map[[20]byte]bool, len(first))
		for f := range first {
			matched[f] = true
		}

		// Delete non-intersecting matches
		for i := 1; i < len(regular); i++ {
			files := tagIndex[string(regular[i])]
			if files == nil {
				return
			}
			for f := range matched {
				if !files[f] {
					delete(matched, f)
				}
			}
		}

		if len(matched) == 0 {
			return
		}
	}

	if len(system) != 0 {
		matched, err = searchBySystemTags(matched, system)
		if err != nil {
			return
		}
	}

	return
}
