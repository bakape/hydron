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
	"log"
	"net/http"
	_ "net/http/pprof"
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

// Synchronise with hydrus tag repository
func syncToHydrus() (err error) {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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

	// File SHA256 hashes, who's tags have been modified
	modified := make(map[[32]byte]struct{}, 1<<10)

	// Distribute update hashes to process across workers.
	// Keep the workers no more than distBuffer ahead of the serializer to
	// conserve memory.
	for i := 0; i < len(meta.Hashes) && i < distBuffer; i++ {
		distribute <- request{i, meta.Hashes[i]}
	}

	// Serialize back to retain update ordering
	for i, ch := range parsed {
		select {
		case err = <-errCh:
			return
		case update := <-ch:
			// Pass next request to the workers
			i += distBuffer
			if i < len(meta.Hashes) {
				distribute <- request{i, meta.Hashes[i]}
			}

			// Commit update to database
			var mod [][32]byte
			mod, err = update.Commit()
			if err != nil {
				return
			}
			for _, m := range mod {
				modified[m] = struct{}{}
			}
			p.done++
			p.print()
		}
	}

	// Apply tags to any matching imported images
	p.close()
	p = progressLogger{
		header: "applying tags",
		total:  len(modified),
	}

	tagMu.Lock()
	defer tagMu.Unlock()

	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	sha256Buc := tx.Bucket([]byte("sha256"))
	imageBuc := tx.Bucket([]byte("images"))
	for sha256 := range modified {
		// Not imported?
		sha1 := sha256Buc.Get(sha256[:])
		if sha1 == nil {
			p.advance()
			continue
		}

		// Was deleted or Hydrus inconsistency?
		img := record(imageBuc.Get(sha1))
		tags := mapHydrusTags(tx, sha256)
		if img == nil || tags == nil {
			p.advance()
			continue
		}

		err = writeRecordTx(
			tx,
			keyValue{
				sha1:   extractKey(sha1),
				record: img,
			},
			tags,
		)
		if err != nil {
			return
		}

		p.advance()
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
	return
}

// Map hydrus tags from tag IDs associated with the SHA256 hash
func mapHydrusTags(tx *bolt.Tx, sha256 [32]byte) [][]byte {
	buc := tx.Bucket([]byte("hydrus"))
	buf := buc.Bucket([]byte("hash->tags")).Get(sha256[:])
	if buf == nil {
		return nil
	}

	l := len(buf) / 8
	tagBuc := buc.Bucket([]byte("tags"))
	tags := make([][]byte, 0, l)
	for i := 0; i < l; i++ {
		tag := tagBuc.Get(buf[i*8 : (i+1)*8])
		if tag != nil {
			tags = append(tags, tag)
		}
	}
	return tags
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
