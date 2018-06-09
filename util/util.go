package util

import (
	"fmt"
	"strings"
)

// Returns, if URL shoiuld be fetched over network
func IsFetchable(url string) bool {
	for _, pre := range [...]string{
		"http://", "https://", "ftp://", "ftps://",
	} {
		if strings.HasPrefix(url, pre) {
			return true
		}
	}
	return false
}

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
