// Temporary bindings for child packages.
// TODO: Restructure

package common

import (
	"github.com/boltdb/bolt"
)

var (
	DB           *bolt.DB
	NormalizeTag func([]byte) []byte
)
