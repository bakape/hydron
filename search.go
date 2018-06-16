package main

import (
	"fmt"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/files"
)

// Returns file paths that match params and print to console
// TODO: Random ordering and select one
func searchImages(params string, page int, random bool) error {
	_, err := db.SearchImages(params, page, func(i common.CompactImage) error {
		fmt.Println(files.SourcePath(i.SHA1, i.Type))
		return nil
	})
	return err
}
