package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/bakape/hydron/core"
)

var stderr = log.New(os.Stderr, "", 0)

// Helper for logging progress of an action to the CLI
type progressLogger struct {
	done, total int
	header      string
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

func (p *progressLogger) Done() {
	p.done++
	p.print()
}

func (p *progressLogger) Err(err error) {
	stderr.Printf("\n%s\n", err)
}

func (p *progressLogger) Close() {
	stderr.Print("\n")
}

func (p *progressLogger) SetTotal(n int) {
	p.total = n
}

// Extract a file's SHA1 id from a string input
func stringToSHA1(s string) (id [20]byte, err error) {
	if len(s) != 40 {
		err = core.InvalidIDError(s)
		return
	}
	buf, err := hex.DecodeString(s)
	if err != nil {
		err = core.InvalidIDError(s + " : " + err.Error())
		return
	}

	id = core.ExtractKey(buf)
	return
}
