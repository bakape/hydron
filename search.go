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

	matched, err := searchByTags(splitTagString(tags))
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

	err = iterateRecords(func(k []byte, r record) {
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
