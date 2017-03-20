package main

import (
	"encoding/binary"
	"runtime"

	"github.com/bakape/hydron/hydrus"
	"github.com/boltdb/bolt"
)

// Synchronise with hydrus tag repository. Not safe for concurrent operations
// on the database.
func syncToHydrus() (err error) {
	session, err := hydrus.GetSession()
	if err != nil {
		return
	}

	// Get current update counter
	var ctr uint64
	err = db.View(func(tx *bolt.Tx) error {
		res := tx.Bucket([]byte("meta")).Get([]byte(hydrus.RepoCounterKey))
		if res != nil {
			ctr = binary.LittleEndian.Uint64(res)
		}
		return nil
	})

	meta, err := hydrus.GetRepoMeta(session, ctr)
	if err != nil {
		return
	}

	// Fetch and parse in parallel

	type request struct {
		i    int
		hash string
	}

	parsed := make([]chan hydrus.Committer, len(meta.Hashes))
	for i := range parsed {
		parsed[i] = make(chan hydrus.Committer)
	}

	distribute := make(chan request)
	errCh := make(chan error)
	closeCh := make(chan struct{})
	defer close(closeCh)

	go func() {
		defer close(distribute)
		for i, h := range meta.Hashes {
			select {
			case <-closeCh:
				return
			case distribute <- request{i, h}:
			}
		}
	}()

	// Don't fsync on every transaction for better performance
	db.NoSync = true
	defer func() {
		db.NoSync = false
	}()

	// Worker count needs to be at least 8, to amortize for request time and
	// latency
	n := runtime.NumCPU() * 2
	if n < 8 {
		n = 8
	}
	for i := 0; i < n; i++ {
		go func() {
			for req := range distribute {
				// Exit on error elsewhere
				select {
				case <-closeCh:
					return
				default:
				}

				com, err := hydrus.GetUpdate(req.hash, session)
				if err != nil {
					select {
					case errCh <- err:
					case <-closeCh:
					}
					return
				}

				// This is like a cancelable promise
				i := req.i
				go func() {
					select {
					case parsed[i] <- com:
					case <-closeCh:
					}
				}()
			}
		}()
	}

	p := progressLogger{
		header: "syncing tags from hydrus",
		total:  len(meta.Hashes),
	}
	defer p.close()

	var (
		tx          *bolt.Tx
		keysWritten int
	)

	// Commit transaction each and 100000 keys, to reduce memory usage
	maybeCommit := func() error {
		keysWritten++
		if keysWritten < 100000 {
			return nil
		}

		keysWritten = 0
		err = tx.Commit()
		if err != nil {
			return err
		}
		tx, err = db.Begin(true)
		return err
	}

	// BoltDB is slow at random writes, so presort the keys before writing to DB
	hashSorter, err := hydrus.NewSorter(40, func(buf []byte) (err error) {
		err = tx.Bucket([]byte("hydrus")).
			Bucket([]byte("hash->id")).
			Put(buf[:32], buf[32:])
		if err != nil {
			return
		}

		p.advance()
		return maybeCommit()
	})
	if err != nil {
		return
	}
	hashTagSorter, err := hydrus.NewSorter(16, func(buf []byte) (err error) {
		err = tx.Bucket([]byte("hydrus")).
			Bucket([]byte("hash:tag")).
			Put(buf, nil)
		if err != nil {
			return
		}

		p.advance()
		return maybeCommit()
	})
	if err != nil {
		return
	}

	// Serialize tag definition updates and write hash definitions and
	// hash->tag relations to external sorters
	for _, ch := range parsed {
		select {
		case err = <-errCh:
			return
		case com := <-ch:
			if com != nil {
				err = com.Commit(hashSorter, hashTagSorter)
				if err != nil {
					return
				}
			}
			p.advance()
		}
	}

	tagMu.Lock()
	defer tagMu.Unlock()

	tx, err = db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	p.close()
	p = progressLogger{
		header: "writing file hashes",
	}
	err = hashSorter.Close()
	if err != nil {
		return
	}

	p.close()
	p = progressLogger{
		header: "writing file->tag relations",
	}
	err = hashTagSorter.Close()
	if err != nil {
		return
	}

	// Apply tags to any matching imported images
	p.close()
	sha256Buc := tx.Bucket([]byte("sha256"))
	p = progressLogger{
		header: "applying tags",
		total:  sha256Buc.Stats().KeyN,
	}
	err = iterateRecordsTx(tx, func(k []byte, r record) error {
		defer p.advance()

		tags := hydrus.MapTags(tx, sha256Buc.Get(k))
		if tags == nil {
			return nil
		}

		return writeRecordTx(
			tx,
			keyValue{
				sha1:   extractKey(k),
				record: r,
			},
			tags,
		)
	})
	if err != nil {
		return
	}

	// Write new counter to database
	err = tx.Bucket([]byte("meta")).Put(
		[]byte(hydrus.RepoCounterKey),
		hydrus.EncodeUint64(ctr+uint64(meta.Count)),
	)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}
	return db.Sync() // Sync changes to disk
}
