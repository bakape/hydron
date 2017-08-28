package main

import "github.com/bakape/hydron/core"

type fetchLogger struct {
	progressLogger
	fetched int
}

func (f *fetchLogger) Done(kv core.KeyValue) {
	f.fetched++
	f.progressLogger.Done(kv)
}

func (f *fetchLogger) NoMatch(kv core.KeyValue) {
	f.progressLogger.Done(kv)
}

func (f *fetchLogger) Close() {
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
