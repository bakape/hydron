package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"

	"github.com/bakape/hydron/core"
)

// Returns file paths that match all tags and print to console
func searchPathsByTags(tags string, random bool) (err error) {
	if tags == "" {
		return printAllPaths(random)
	}

	matched, err := core.SearchByTags(core.SplitTagString(tags, ' '))
	if err != nil || len(matched) == 0 {
		return
	}
	types := make(map[[20]byte]core.FileType, len(matched))

	// Retrieve file types of all matches
	err = core.MapRecords(matched, func(id [20]byte, r core.Record) {
		types[id] = r.Type()
	})
	if err != nil {
		return
	}

	// Convert to paths
	for id := range matched {
		printPath(id[:], types[id])
		if random {
			return
		}
	}

	return
}

func printPath(id []byte, typ core.FileType) {
	fmt.Println(core.SourcePath(hex.EncodeToString(id), typ))
}

// Print paths to all files
func printAllPaths(random bool) (err error) {
	ids := make([][]byte, 0, 1<<10)
	types := make([]core.FileType, 0, 1<<10)

	err = core.IterateRecords(func(k []byte, r core.Record) {
		ids = append(ids, k)
		types = append(types, r.Type())
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
