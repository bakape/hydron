package main

import (
	"encoding/hex"
	"fmt"
)

// Returns file paths that match all tags and print to console
func searchPathsByTags(tags string) (err error) {
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
	for i, id := range matched {
		fmt.Println(sourcePath(hex.EncodeToString(id[:]), types[i]))
	}

	return nil
}
