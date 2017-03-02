package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	fetchTags  = make(chan tagFetchRequest)
	errNoMatch = errors.New("no match found")
)

type tagFetchRequest struct {
	md5 [16]byte
	res chan<- tagFetchResponse
}

type tagFetchResponse struct {
	md5  [16]byte
	tags [][]byte
	err  error
}

// Spawn 4 goroutines to perform parallel fetches
func init() {
	for i := 0; i < 4; i++ {
		go func() {
			for {
				req := <-fetchTags
				tags, err := fetchFromGelbooru(req.md5)
				req.res <- tagFetchResponse{
					md5:  req.md5,
					tags: tags,
					err:  err,
				}
			}
		}()
	}
}

func fetchFromGelbooru(md5 [16]byte) (tags [][]byte, err error) {
	r, err := http.Get(gelbooruURL(md5))
	if err != nil {
		return
	}
	defer r.Body.Close()

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	if len(buf) == 0 {
		err = errNoMatch
		return
	}

	type Data []struct {
		Rating, Tags string
	}
	var data Data
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return
	}

	if len(data) == 0 {
		err = errNoMatch
		return
	}
	d := data[0]
	if len(d.Tags) == 0 {
		err = errNoMatch
		return
	}

	tags = splitTagString(d.Tags, ' ')
	if d.Rating != "" {
		r := d.Rating[0]
		if r != ' ' && r != '\x00' {
			new := make([][]byte, len(tags)+1)
			copy(new, tags)
			new[len(tags)] = []byte{'r', 'a', 't', 'i', 'n', 'g', ':', r}
			tags = new
		}
	}

	return
}

func gelbooruURL(md5 [16]byte) string {
	return fmt.Sprintf(
		"http://gelbooru.com/index.php?page=dapi&s=post&q=index&tags=md5:%s&json=1",
		hex.EncodeToString(md5[:]),
	)
}

// Attempt to fetch tags for all files that have not yet had their tags synced
// with gelbooru.com
func fetchAllTags() error {
	// Get list of candidate files
	files := make(map[[16]byte]keyValue, 64)
	err := iterateRecords(func(k []byte, r record) {
		if canFetchTags(r) && !r.HaveFetchedTags() {
			files[r.MD5()] = keyValue{
				sha1:   extractKey(k),
				record: r.Clone(),
			}
		}
	})
	if err != nil {
		return err
	}

	return fetchFileTags(files)
}

// Return, if file is eligible for fetching tags from boorus
func canFetchTags(r record) bool {
	switch r.Type() {
	case jpeg, png, gif, webm:
		return true
	}
	return false
}

// Fetch tags for listed files
func fetchFileTags(files map[[16]byte]keyValue) error {
	res := make(chan tagFetchResponse)
	go func() {
		for md5 := range files {
			fetchTags <- tagFetchRequest{
				md5: md5,
				res: res, // Simply fan-in the response
			}
		}
	}()

	p := progressLogger{
		total:  len(files),
		header: "fetching tags",
	}
	defer p.close()
	for range files {
		switch r := <-res; r.err {
		case nil:
			err := writeFetchedTags(files[r.md5], r.tags)
			if err != nil {
				return err
			}
			p.done++
		case errNoMatch:
		default:
			p.printError(r.err)
		}
		p.print()
	}

	return nil
}

func writeFetchedTags(kv keyValue, tags [][]byte) error {
	kv.SetFetchedTags()
	return writeRecord(kv, tags)
}

func fetchSingleFileTags(kv keyValue) (err error) {
	res := make(chan tagFetchResponse)
	fetchTags <- tagFetchRequest{
		md5: kv.MD5(),
		res: res,
	}
	r := <-res
	switch r.err {
	case nil:
		err = writeFetchedTags(kv, r.tags)
	case errNoMatch:
		err = nil
	}
	return
}
