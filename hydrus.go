package main

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

const (
	repoAddr       = "https://hydrus.no-ip.org:45871"
	repoKey        = "4a285629721ca442541ef2c15ea17d1f7f7578b0c3f4f5f2a05f8f0ab297786f"
	repoCounterKey = "hydrus_repo_counter:default"
)

// Hydrus uses a self-signed cert. Must bypass the regular security checks.
var hydrusClient = http.Client{
	Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// Synchronise with hydrus tag repository. Not safe for concurrent operations
// on the database
func syncToHydrus() (err error) {
	session, err := getSession()
	if err != nil {
		return
	}

	// Get current update counter
	var ctr uint64
	err = db.View(func(tx *bolt.Tx) error {
		res := tx.Bucket([]byte("meta")).Get([]byte(repoCounterKey))
		if res != nil {
			ctr = binary.LittleEndian.Uint64(res)
		}
		return nil
	})

	meta, err := getRepoMeta(session, ctr)
	if err != nil {
		return
	}

	// Fetch and parse on 4 threads in parallel
	type request struct {
		i    int
		hash string
	}

	const distBuffer = 30
	distribute := make(chan request, distBuffer)
	errCh := make(chan error, 4)
	closeCh := make(chan struct{})
	parsed := make([]chan hydrusUpdate, len(meta.Hashes))
	for i := range parsed {
		parsed[i] = make(chan hydrusUpdate)
	}
	defer close(closeCh)
	defer close(distribute)

	// Workers
	for i := 0; i < 8; i++ {
		go func() {
			for req := range distribute {
				// Exit on error elsewhere
				select {
				case <-closeCh:
					return
				default:
				}

				update, err := getHydrusUpdate(req.hash, session)
				if err != nil {
					errCh <- err
					return
				}

				// This acts as a cancelable promise
				i := req.i
				go func() {
					select {
					case <-closeCh: // To prevent goroutine leaks
					case parsed[i] <- update:
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

	// Distribute update hashes to process across workers.
	// Keep the workers no more than distBuffer ahead of the serializer to
	// conserve memory.
	for i := 0; i < len(meta.Hashes) && i < distBuffer; i++ {
		distribute <- request{i, meta.Hashes[i]}
	}

	// Don't fsync on every transaction for better performance
	db.NoSync = true
	defer func() {
		db.NoSync = false
	}()

	// Serialize back to retain update ordering
	for i, ch := range parsed {
		select {
		case err = <-errCh:
			return
		case update := <-ch:
			// Pass next request to the workers
			j := i + distBuffer
			if j < len(meta.Hashes) {
				distribute <- request{j, meta.Hashes[j]}
			}

			// Commit update to database
			err = update.Commit()
			if err != nil {
				return
			}

			// Sync to disk every 100 commits
			if i != 0 && i%100 == 0 {
				err = db.Sync()
				if err != nil {
					return
				}
			}

			p.advance()
		}
	}

	tagMu.Lock()
	defer tagMu.Unlock()

	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	// Apply tags to any matching imported images
	p.close()
	sha256Buc := tx.Bucket([]byte("sha256"))
	p = progressLogger{
		header: "applying tags",
		total:  sha256Buc.Stats().KeyN,
	}
	err = iterateRecordsTx(tx, func(k []byte, r record) error {
		defer p.advance()

		tags := mapHydrusTags(tx, sha256Buc.Get(k))
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
		[]byte(repoCounterKey),
		encodeUint64(ctr+uint64(meta.Count)),
	)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}
	err = db.Sync() // Sync changes to disk
	return
}

// Map hydrus tags from tag IDs associated with the SHA256 hash
func mapHydrusTags(tx *bolt.Tx, sha256 []byte) (tags [][]byte) {
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

// Get list of updated files
func getRepoMeta(session *http.Cookie, ctr uint64) (meta repoMeta, err error) {
	url := repoAddr + "/metadata?since=" + strconv.FormatUint(ctr, 10)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.AddCookie(session)
	res, err := hydrusClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	buf, err := decodeZlib(res)
	if err != nil {
		return
	}
	err = meta.UnMarshalJSON(buf)
	return
}

func decodeZlib(res *http.Response) (buf []byte, err error) {
	var w bytes.Buffer

	zr, err := zlib.NewReader(res.Body)
	if err != nil {
		return
	}
	defer zr.Close()

	_, err = io.Copy(&w, zr)
	buf = w.Bytes()
	return
}

// Get session key to use for connections
func getSession() (session *http.Cookie, err error) {
	req, err := http.NewRequest("GET", repoAddr+"/session_key", nil)
	if err != nil {
		return
	}
	req.Header.Set("Hydrus-Key", repoKey)
	res, err := hydrusClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	for _, c := range res.Cookies() {
		if c.Name == "session_key" {
			session = c
			break
		}
	}
	if session == nil {
		err = errors.New("no session key received")
	}
	return
}

// Fetch and parse a new Hydrus tag repo update
func getHydrusUpdate(id string, session *http.Cookie) (
	update hydrusUpdate, err error,
) {
	url := repoAddr + "/update?update_hash=" + id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.AddCookie(session)
	res, err := hydrusClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	buf, err := decodeZlib(res)
	if err != nil {
		return
	}

	// Initial decode and determine update type
	var arr []json.RawMessage
	err = json.Unmarshal(buf, &arr)
	if err != nil {
		return
	}
	if len(arr) < 3 {
		err = invalidStructureError(buf)
		return
	}
	var typ uint
	err = json.Unmarshal(arr[0], &typ)
	if err != nil {
		return
	}
	switch typ {
	case 36:
		update = new(definitionUpdate)
	case 34:
		update = new(contentUpdate)
	default:
		err = fmt.Errorf("unknown update type: %d", typ)
		return
	}

	err = update.Parse(arr)
	return
}
