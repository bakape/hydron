package main

import (
	"strings"

	"github.com/bakape/hydron/core"
)

// Remove files from the database by ID from the CLI
func removeFiles(ids []string) error {
	for _, id := range ids {
		sha1, err := stringToSHA1(id)
		if err != nil {
			return err
		}
		err = core.RemoveFile(sha1)
		if err != nil {
			return err
		}
	}
	return nil
}

// Add tags to the target file from the CLI
func addTags(id string, tags []string) error {
	return modTags(id, tags, core.AddTags)
}

func modTags(
	id string,
	tags []string,
	fn func([20]byte, [][]byte) error,
) error {
	sha1, err := stringToSHA1(id)
	if err != nil {
		return err
	}
	return fn(sha1, core.SplitTagString(strings.Join(tags, " "), ' '))
}

// Remove tags from the target file from the CLI
func removeTags(id string, tags []string) error {
	return modTags(id, tags, core.RemoveTags)
}
