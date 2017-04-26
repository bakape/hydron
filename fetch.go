package main

import "github.com/bakape/hydron/core"

type fetchLogger struct {
	progressLogger
	fetched int
}

func (f *fetchLogger) Done() {
	f.fetched++
	f.progressLogger.Done()
}

func (f *fetchLogger) NoMatch() {
	f.progressLogger.Done()
}

func (f *fetchLogger) Close() {
	f.progressLogger.Close()
	if f.total != f.fetched {
		stderr.Printf("no tags found for %d files", f.total-f.fetched)
	}
}

func fetchAllTags() error {
	return core.FetchAllTags(&fetchLogger{
		progressLogger: progressLogger{
			header: "fetching tags",
		},
	})
}
