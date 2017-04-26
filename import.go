package main

import "github.com/bakape/hydron/core"

func importPaths(paths []string, del, fetchTags bool, tags string) (err error) {
	var t [][]byte
	if tags != "" {
		t = core.SplitTagString(tags, ' ')
	}

	iLogger := progressLogger{
		header: "importing and thumbnailing",
	}
	fLogger := fetchLogger{
		progressLogger: progressLogger{
			header: "fetching tags",
		},
	}
	return core.ImportPaths(paths, del, fetchTags, t, &iLogger, &fLogger)
}
