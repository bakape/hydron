package core

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

// FetchLogger provides callbacks for displaying tag fetching progress
type FetchLogger interface {
	Logger
	NoMatch()
}

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

	tags = SplitTagString(d.Tags, ' ')
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

// FetchAllTags attempts to fetch tags for all files that have not yet had their
// tags synced with gelbooru.com
func FetchAllTags(logger FetchLogger) error {
	// Get list of candidate files
	files := make(map[[16]byte]KeyValue, 64)
	err := IterateRecords(func(k []byte, r Record) {
		if CanFetchTags(r) && !r.HaveFetchedTags() {
			files[r.MD5()] = KeyValue{
				SHA1:   ExtractKey(k),
				Record: r.Clone(),
			}
		}
	})
	if err != nil {
		return err
	}

	return FetchTags(files, logger)
}

// Return, if file is eligible for fetching tags from boorus
func CanFetchTags(r Record) bool {
	switch r.Type() {
	case JPEG, PNG, GIF, WEBM:
		return true
	}
	return false
}

//  FetchTags fetches tags for each of files
func FetchTags(files map[[16]byte]KeyValue, l FetchLogger) error {
	res := make(chan tagFetchResponse)
	go func() {
		for md5 := range files {
			fetchTags <- tagFetchRequest{
				md5: md5,
				res: res, // Simply fan-in the response
			}
		}
	}()

	l.SetTotal(len(files))
	defer l.Close()
	for range files {
		switch r := <-res; r.err {
		case nil:
			if r.tags != nil {
				if err := writeFetchedTags(files[r.md5], r.tags); err != nil {
					return err
				}
			}
			l.Done()
		case errNoMatch:
			l.NoMatch()
		default:
			l.Err(r.err)
		}
	}

	return nil
}

func writeFetchedTags(kv KeyValue, tags [][]byte) error {
	kv.setFetchedTags()
	return writeRecord(kv, tags)
}

// Fetch tags of a single file
func FetchSingleFileTags(kv KeyValue) (err error) {
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
