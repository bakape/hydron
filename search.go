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
	matched := searchByTags(splitTagString(tags))
	types := make([]fileType, len(matched))

	// Retrieve file types of all tags
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
		fmt.Println(sourcePath(hex.EncodeToString(matched[i][:]), types[i]))
	}
	if random {
		print(rand.Intn(len(matched)))
	} else {
		for i := range matched {
			print(i)
		}
	}

	return nil
}
