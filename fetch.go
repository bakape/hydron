package main

import (
	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/fetch"
)

// Fetch and update tags for all stored images
func fetchAllTags() error {
	all, err := db.GetGelbooruTaggable()
	if err != nil {
		return err
	}

	// Buffer all into a channel
	passAll := make(chan db.IDAndMD5, len(all))
	for _, img := range all {
		passAll <- img
	}

	// Process images in parallel
	ch := make(chan error)
	for i := 0; i < fetch.FetcherCount; i++ {
		go func() {
			for img := range passAll {
				tags, changed, err := fetch.FetchTags(img.MD5)
				if err != nil {
					goto end
				}
				err = db.UpdateTags(img.ID, tags, common.Gelbooru, changed)
			end:
				ch <- err
			}
		}()
	}

	// Aggregate and log results
	p := progressLogger{
		header: "fetching tags",
		total:  len(all),
	}
	for i := 0; i < len(all); i++ {
		err = <-ch
		if err != nil {
			p.Err(err)
		} else {
			p.Done()
		}
	}
	close(passAll) // Stop worker goroutines
	p.Close()

	return nil
}
