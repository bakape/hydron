package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

var stderr = log.New(os.Stderr, "", 0)

// User passed an invalid file ID
type invalidIDError string

func (e invalidIDError) Error() string {
	return "invalid id: " + string(e)
}

// Helper for logging progress of an action to the CLI
type progressLogger struct {
	done, total int
	header      string
	finalize    func(*progressLogger)
}

func (p *progressLogger) print() {
	fmt.Fprintf(
		os.Stderr,
		"\r%s: %d / %d - %.2f%%",
		p.header,
		p.done, p.total,
		float32(p.done)/float32(p.total)*100,
	)
}

func (p *progressLogger) printError(err error) {
	stderr.Printf("\n%s\n", err)
}

func (p *progressLogger) close() {
	stderr.Print("\n")
	if p.finalize != nil {
		p.finalize(p)
	}
}

// Extract a copy of the underlying SHA1 key from a BoltDB key
func extractKey(k []byte) (sha1 [20]byte) {
	for i := range sha1 {
		sha1[i] = k[i]
	}
	return
}

// Extract a file's SHA1 id from a string input
func stringToSHA1(s string) (id [20]byte, err error) {
	if len(s) != 40 {
		err = invalidIDError(s)
		return
	}
	buf, err := hex.DecodeString(s)
	if err != nil {
		err = invalidIDError(s + " : " + err.Error())
		return
	}

	id = extractKey(buf)
	return
}

// Attach a descriptive prefix to an existing error
func wrapError(err error, format string, args ...interface{}) error {
	return fmt.Errorf("%s: %s", fmt.Sprintf(format, args...), err)
}
