package util

import (
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
