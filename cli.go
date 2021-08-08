package main

import (
	"fmt"

	"github.com/bakape/hydron/v3/common"
	"github.com/bakape/hydron/v3/db"
	"github.com/bakape/hydron/v3/files"
	"github.com/bakape/hydron/v3/tags"
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
	return db.AddTags(id, tags.FromString(tagStr, common.User))
}

// Remove tags from the target file from the CLI
func removeTags(sha1 string, tagStr string) error {
	id, err := db.GetImageID(sha1)
	if err != nil {
		return err
	}
	return db.RemoveTags(id, tags.FromString(tagStr, common.User))
}

// Set the target file's name from the CLI
func setImageName(sha1 string, name string) error {
	id, err := db.GetImageID(sha1)

	if err != nil {
		return err
	}

	return db.SetName(id, name)
}

// Returns file paths that match params and print to console
func searchImages(params string) (err error) {
	var page common.Page
	err = tags.ParseFilters(params, &page)
	if err != nil {
		return
	}
	return db.SearchImages(&page, false, func(i common.CompactImage) error {
		fmt.Println(files.SourcePath(i.SHA1, i.Type))
		return nil
	})
}
