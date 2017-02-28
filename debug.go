package main

import (
	"encoding/hex"
	"fmt"
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
	return iterateRecords(func(k []byte, r record) {
		md5 := r.MD5()
		tags := r.Tags()
		rr := readableRecord{
			ext:        extensions[r.Type()],
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
		rr.bools[0] = boolToUint8(r.HaveFetchedTags())
		rr.bools[1] = boolToUint8(r.ThumbIsPNG())
		fmt.Printf("> %s: %v\n", hex.EncodeToString(k), rr)
	})
}

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
