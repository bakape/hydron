package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/boltdb/bolt"
)

const dbVersion = 1

var dbUpgrades = map[uint16]func(*bolt.Tx) error{
	// Calculate SHA256 hashes for all file.
	// Needed for interoperation with Hydrus.
	0: func(tx *bolt.Tx) (err error) {
		p := progressLogger{
			header: "upgrading database",
			total:  tx.Bucket([]byte("images")).Stats().KeyN,
		}
		defer p.close()

		err = iterateRecordsTx(tx, func(k []byte, r record) (err error) {
			f, err := os.Open(sourcePath(hex.EncodeToString(k), r.Type()))
			if err != nil {
				return
			}
			defer f.Close()

			h := sha256.New()
			_, err = io.Copy(h, f)
			if err != nil {
				return
			}
			err = putSHA256(tx, k, h.Sum(nil))
			if err != nil {
				return
			}

			p.done++
			p.print()
			return nil
		})
		if err != nil {
			return
		}

		return
	},
}

// Check the database version and perform any needed upgrades
func checkVersion(tx *bolt.Tx) (err error) {
	var ver uint16
	if res := tx.Bucket([]byte("meta")).Get([]byte("version")); res != nil {
		ver = binary.LittleEndian.Uint16(res)
	}

	if ver > dbVersion {
		return fmt.Errorf("incompatible database version: %d", ver)
	}
	for i := ver; i < dbVersion; i++ {
		err = dbUpgrades[i](tx)
		if err != nil {
			return
		}

		// Write new version number
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, i+1)
		err = tx.Bucket([]byte("meta")).Put([]byte("version"), buf)
		if err != nil {
			return
		}
	}

	return
}
