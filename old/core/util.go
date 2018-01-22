package core

import (
	"encoding/hex"
	"fmt"
)

// User passed an invalid file ID
type InvalidIDError string

func (e InvalidIDError) Error() string {
	return "invalid id: " + string(e)
}

type sha1Sorter [][20]byte

func (s sha1Sorter) Len() int {
	return len(s)
}

func (s sha1Sorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sha1Sorter) Less(i, j int) bool {
	a, b := s[i], s[j]
	for i := range a {
		switch {
		case a[i] < b[i]:
			return true
		case a[i] > b[i]:
			return false
		}
	}
	return true
}

// Extract a copy of the underlying SHA1 key from a BoltDB key
func ExtractKey(k []byte) (sha1 [20]byte) {
	for i := range sha1 {
		sha1[i] = k[i]
	}
	return
}

// Attach a descriptive prefix to an existing error
func wrapError(err error, format string, args ...interface{}) error {
	return fmt.Errorf("%s: %s", fmt.Sprintf(format, args...), err)
}

// Waterfall executes a slice of functions until the first error returned. This
// error, if any, is returned to the caller.
func Waterfall(fns ...func() error) (err error) {
	for _, fn := range fns {
		err = fn()
		if err != nil {
			break
		}
	}
	return
}

// Extract a file's SHA1 id from a string input
func StringToSHA1(s string) (id [20]byte, err error) {
	if len(s) != 40 {
		err = InvalidIDError(s)
		return
	}
	buf, err := hex.DecodeString(s)
	if err != nil {
		err = InvalidIDError(s + " : " + err.Error())
		return
	}

	id = ExtractKey(buf)
	return
}
