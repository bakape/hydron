package hydrus

import (
	"bytes"

	"github.com/boltdb/bolt"
)

// MapTags maps hydrus tags from tag IDs associated with the SHA256 hash
func MapTags(tx *bolt.Tx, sha256 []byte) (tags [][]byte) {
	buc := tx.Bucket([]byte("hydrus"))
	id := buc.Bucket([]byte("hash->id")).Get(sha256)
	if id == nil {
		return
	}

	tags = make([][]byte, 0, 16)
	c := buc.Bucket([]byte("hash:tag")).Cursor()
	buc = buc.Bucket([]byte("id->tag"))
	for k, _ := c.Seek(id); k != nil && bytes.HasPrefix(k, id); k, _ = c.Next() {
		tag := buc.Get(k[8:])
		if tag != nil {
			tags = append(tags, tag)
		}
	}
	return
}
