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
				req.res <- tagFetchResponse{tags, err}
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

	tags = splitTagString(d.Tags)
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
	files := make(map[[20]byte]record, 64)
	err := iterateRecords(func(k []byte, r record) {
		switch r.Type() {
		case jpeg, png, gif, webm:
			if !r.HaveFetchedTags() {
				files[extractKey(k)] = r.Clone()
			}
		}
	})
	if err != nil {
		return err
	}

	return fetchFileTags(files)
}

// Fetch tags for listed files
func fetchFileTags(files map[[20]byte]record) error {
	res := make(chan tagFetchResponse)
	go func() {
		for _, f := range files {
			fetchTags <- tagFetchRequest{
				md5: f.MD5(),
				res: res, // Simply fan-in the response
			}
		}
	}()

	p := progressLogger{
		total:  len(files),
		header: "fetching tags",
	}
	defer p.close()
	for k, v := range files {
		switch r := <-res; r.err {
		case nil:
			v.SetFetchedTags()
			kv := keyValue{
				SHA1:   k,
				record: v,
			}
			err := writeRecord(kv, mergeTagSets(v.Tags(), r.tags))
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
