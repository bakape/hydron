package fetch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/tags"
)

var fetchTags = make(chan request)

// Number of parallel fetch workers
const FetcherCount = 8

type request struct {
	md5 string
	res chan<- response
}

type response struct {
	lastChange int64
	tags       []common.Tag
	err        error
}

// Fetch tags from gelbooru for a given MD5 hash.
// Return tags and last change time.
func FetchTags(md5 string) ([]common.Tag, time.Time, error) {
	ch := make(chan response)
	fetchTags <- request{md5, ch}
	res := <-ch
	return res.tags, time.Unix(res.lastChange, 0), res.err
}

// Return, if fileType can possibly have tags fetched
func CanFetchTags(t common.FileType) bool {
	switch t {
	case common.JPEG, common.PNG, common.GIF, common.WEBM:
		return true
	default:
		return false
	}
}

// Spawn 8 goroutines to perform parallel fetches
func init() {
	for i := 0; i < FetcherCount; i++ {
		go func() {
			for {
				req := <-fetchTags
				tags, lastChange, err := fetchFromGelbooru(req.md5)
				req.res <- response{
					tags:       tags,
					lastChange: lastChange,
					err:        err,
				}
			}
		}()
	}
}

func fetchFromGelbooru(md5 string) (
	re []common.Tag, lastChange int64, err error,
) {
	url := fmt.Sprintf(
		"http://gelbooru.com/index.php?page=dapi&s=post&q=index&tags=md5:%s&json=1",
		md5)
	r, err := http.Get(url)
	if err != nil {
		return
	}
	defer r.Body.Close()

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil || len(buf) == 0 {
		return
	}

	type Data []struct {
		Change       uint64
		Rating, Tags string
	}
	var data Data
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return
	}

	if len(data) == 0 {
		return
	}
	d := data[0]
	if len(d.Tags) == 0 {
		return
	}

	re = tags.FromString(d.Tags+" rating:"+d.Rating, common.Gelbooru)
	lastChange = int64(d.Change)
	return
}
