package core

// Init prepares hydron for operation
func Init() error {
	return Waterfall(initDirs, openDB)
}

// ShutDown gracefully shuts down the core runtime
func ShutDown() {
	db.Close()
}
