package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Returns file paths that match all tags and print to console
func searchPathsByTags(tags string, random bool) (err error) {
	if tags == "" {
		return printAllPaths(random)
	}

	matched, err := searchByTags(splitTagString(tags, ' '))
	if err != nil {
		return
	}
	types := make([]fileType, len(matched))

	// Retrieve file types of all matches
	tx, err := db.Begin(false)
	if err != nil {
		return
	}
	buc := tx.Bucket([]byte("images"))
	for i, id := range matched {
		types[i] = record(buc.Get(id[:])).Type()
	}
	err = tx.Rollback()
	if err != nil {
		return
	}

	// Convert to paths
	print := func(i int) {
		printPath(matched[i][:], types[i])
	}
	if random {
		print(rand.Intn(len(matched)))
	} else {
		for i := range matched {
			print(i)
		}
	}

	return
}

func printPath(id []byte, typ fileType) {
	fmt.Println(sourcePath(hex.EncodeToString(id), typ))
}

// Print paths to all files
func printAllPaths(random bool) (err error) {
	ids := make([][]byte, 0, 512)
	types := make([]fileType, 0, 512)

	err = iterateRecords(func(k []byte, r record) error {
		ids = append(ids, k)
		types = append(types, r.Type())
		return nil
	})
	if err != nil {
		return
	}

	print := func(i int) {
		printPath(ids[i], types[i])
	}
	if random {
		print(rand.Intn(len(ids)))
	} else {
		for i := range ids {
			print(i)
		}
	}

	return
}

// Find files that match all tags
func searchByTags(tags [][]byte) (arr [][20]byte, err error) {
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

	var matched map[[20]byte]bool

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

	arr = make([][20]byte, 0, len(matched))
	for f := range matched {
		arr = append(arr, f)
	}
	return
}
