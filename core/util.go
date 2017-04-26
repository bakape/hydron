package core

import (
	"fmt"
)

// User passed an invalid file ID
type InvalidIDError string

func (e InvalidIDError) Error() string {
	return "invalid id: " + string(e)
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
