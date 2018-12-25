package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
)

var stderr = log.New(os.Stderr, "", 0)

// Helper for logging progress of an action to the CLI
type progressLogger struct {
	lastWasError bool
	done, total  int
	header       string
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
	p.lastWasError = false
	p.print()
}

func (p *progressLogger) Err(err error) {
	if !p.lastWasError {
		fmt.Fprint(os.Stderr, "\n")
	}
	fmt.Fprintf(os.Stderr, "%s\n", err)
	p.lastWasError = true
}

func (p *progressLogger) Close() {
	fmt.Println("")
}

// Convert import page's input string into string array of file paths
func parsePaths(input string) ([]string, error) {
	var store []string
	inSpace := false
	quote := 0
	prev := 0

	for i, c := range input {
		if (unicode.IsSpace(c)) {
			// If inside quotes or prev. rune was whitespace, ignore whitespace
			if (quote%2 == 1 || inSpace) {
				continue
			}
			inSpace = true
			store = append(store, input[prev:i])
			prev = i + 1
		}else {
			if (c == '"') {
				quote++
			}
			inSpace = false
		}
	}
	if (quote%2 == 1) {
		return store, errors.New("Invalid formatting: uneven quote marks.")
	}
	if !inSpace {
		store = append(store, input[prev:])
	}
	// Have to remove encasing quotes to use file path
	if (quote != 0) {
		for i, s := range store {
			store[i] = strings.Replace(s, `"`, ``, -1)
		}
	}

	return store, nil
}