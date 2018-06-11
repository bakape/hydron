package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/bakape/hydron/files"
	imp "github.com/bakape/hydron/import"
)

/*
Import file/directory paths
paths: list for file and/pr directory paths
del: delete files after import
fetchTags: also fetch tags from gelbooru
tagStr: add string of tags to all imported files
*/
func importPaths(paths []string, del, fetchTags bool, tagStr string) error {
	paths, err := files.Traverse(paths)
	if err != nil {
		return err
	}

	// Buffer all paths into a channel
	passPaths := make(chan string, len(paths))
	for _, p := range paths {
		passPaths <- p
	}

	// Process files in parallel
	ch := make(chan error)
	n := runtime.NumCPU() + 1
	for i := 0; i < n; i++ {
		go func() {
			for p := range passPaths {
				f, err := os.Open(p)
				if err != nil {
					goto end
				}

				_, err = imp.ImportFile(f, tagStr, fetchTags)
				switch err {
				case nil:
				case imp.ErrImported:
					err = nil
				default:
					err = fmt.Errorf("%s: %s", p, err)
					goto end
				}

				if del {
					// Close file before removing
					f.Close()
					f = nil
					err = os.Remove(p)
				}

			end:
				if f != nil {
					f.Close()
				}
				ch <- err
			}
		}()
	}

	// Aggregate and log results
	p := progressLogger{
		header: "importing, thumbnailing",
		total:  len(paths),
	}
	if fetchTags {
		p.header += ", fetching tags"
	}
	for i := 0; i < len(paths); i++ {
		err = <-ch
		if err != nil {
			p.Err(err)
		} else {
			p.Done()
		}
	}
	close(passPaths) // Stop worker goroutines
	p.Close()

	return nil
}
