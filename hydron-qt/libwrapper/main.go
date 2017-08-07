package main

import "C"
import "github.com/bakape/hydron/core"

// Stub
func main() {}

//export startHydron
// Load hydron runtime
func startHydron() *C.char {
	return toCError(core.Init())
}

//export shutDownHydron
// Release any held resources
func shutDownHydron() {
	core.ShutDown()
}
