package main

import (
	"time"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/tags"
)

// Remove files from the database by ID from the CLI
func removeFiles(ids []string) (err error) {
	for _, id := range ids {
		err = db.RemoveImage(id)
		if err != nil {
			return
		}
	}
	return
}

// Add tags to the target file from the CLI
func addTags(sha1 string, tagStr string) error {
	id, err := db.GetImageID(sha1)
	if err != nil {
		return err
	}
	return db.AddTags(id, tags.FromString(tagStr, common.User), time.Now())
}

// Remove tags from the target file from the CLI
func removeTags(sha1 string, tagStr string) error {
	id, err := db.GetImageID(sha1)
	if err != nil {
		return err
	}
	return db.RemoveTags(id, tags.FromString(tagStr, common.User))
}
