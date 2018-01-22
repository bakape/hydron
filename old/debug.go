package main

import (
	"encoding/hex"
	"fmt"

	"github.com/bakape/hydron/core"
)

type readableRecord struct {
	ext              string
	bools            [8]uint8
	importTime, size uint64
	dims             string
	length           uint64
	MD5              string
	tags             []string
}

// Print the entire database in a mostly human-readable format
func printDB() error {
	return core.IterateRecords(func(k []byte, r core.Record) {
		md5 := r.MD5()
		tags := r.Tags()
		rr := readableRecord{
			ext:        core.Extensions[r.Type()],
			importTime: r.ImportTime(),
			size:       r.Size(),
			dims:       fmt.Sprintf("%dx%d", r.Width(), r.Height()),
			length:     r.Length(),
			MD5:        hex.EncodeToString(md5[:]),
			tags:       make([]string, len(tags)),
		}
		for i := range tags {
			rr.tags[i] = string(tags[i])
		}
		for i, fn := range [...]func() bool{
			r.HaveFetchedTags, r.ThumbIsPNG, r.HasThumb,
		} {
			b := uint8(0)
			if fn() {
				b = 1
			}
			rr.bools[i] = b
		}
		fmt.Printf("> %s: %v\n", hex.EncodeToString(k), rr)
	})
}
