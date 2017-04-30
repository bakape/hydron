package main

import "github.com/bakape/hydron/core"

func main() {
	// Load hydron runtime
	if err := core.Init(); err != nil {
		panic(err)
	}
	defer core.ShutDown()
}
