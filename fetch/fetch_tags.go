package fetch

import (
	"errors"

	"github.com/bakape/hydron/common"
)

var (
	fetchTags  = make(chan tagFetchRequest)
	errNoMatch = errors.New("no match found")
)

// FetchLogger provides callbacks for displaying tag fetching progress
type FetchLogger interface {
	common.Logger
	NoMatch(common.Record)
	Close()
}

type tagFetchRequest struct {
	common.Record
	res chan<- tagFetchResponse
}

type tagFetchResponse struct {
	common.Record
	tags []common.Tag
	err  error
}

// // Spawn 4 goroutines to perform parallel fetches
// func init() {
// 	for i := 0; i < 4; i++ {
// 		go func() {
// 			for {
// 				req := <-fetchTags
// 				tags, err := fetchFromGelbooru(req.MD5)
// 				req.res <- tagFetchResponse{
// 					Record: req.Record,
// 					tags:   tags,
// 					err:    err,
// 				}
// 			}
// 		}()
// 	}
// }

// func fetchFromGelbooru(md5 string) (tags []common.Tag, err error) {
// 	r, err := http.Get(gelbooruURL(md5))
// 	if err != nil {
// 		return
// 	}
// 	defer r.Body.Close()

// 	buf, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		return
// 	}
// 	if len(buf) == 0 {
// 		err = errNoMatch
// 		return
// 	}

// 	type Data []struct {
// 		Rating, Tags string
// 	}
// 	var data Data
// 	err = json.Unmarshal(buf, &data)
// 	if err != nil {
// 		return
// 	}

// 	if len(data) == 0 {
// 		err = errNoMatch
// 		return
// 	}
// 	d := data[0]
// 	if len(d.Tags) == 0 {
// 		err = errNoMatch
// 		return
// 	}

// 	tags = taglib.SplitTagString(d.Tags, ' ')
// 	if d.Rating != "" {
// 		r := d.Rating[0]
// 		if r != ' ' && r != '\x00' {
// 			new := make([][]byte, len(tags)+1)
// 			copy(new, tags)
// 			new[len(tags)] = []byte{'r', 'a', 't', 'i', 'n', 'g', ':', r}
// 			tags = new
// 		}
// 	}

// 	return
// }

// func gelbooruURL(md5 string) string {
// 	return fmt.Sprintf(
// 		"http://gelbooru.com/index.php?page=dapi&s=post&q=index&tags=md5:%s&json=1",
// 		md5,
// 	)
// }

// FetchAllTags attempts to fetch tags for all files that have not yet had their
// tags synced with gelbooru.com
func FetchAllTags(logger FetchLogger) error {
	panic("TODO")
	return nil
	// // Get list of candidate files
	// files := make([]KeyValue, 0, 64)
	// err := IterateRecords(func(k []byte, r Record) {
	// 	if CanFetchTags(r) && !r.HaveFetchedTags() {
	// 		files = append(files, KeyValue{
	// 			SHA1:   ExtractKey(k),
	// 			Record: r.Clone(),
	// 		})
	// 	}
	// })
	// if err != nil {
	// 	return err
	// }

	// logger.SetTotal(len(files))

	// // This call is synchronous, but has to use and async function
	// ch := make(chan KeyValue, len(files))
	// go func() {
	// 	for _, f := range files {
	// 		ch <- f
	// 	}
	// 	close(ch)
	// }()
	// err = FetchTags(ch, logger)
	// if err != nil {
	// 	return err
	// }
	// return nil
}

// // Return, if file is eligible for fetching tags from boorus
// func CanFetchTags(r Record) bool {
// 	switch r.Type() {
// 	case JPEG, PNG, GIF, WEBM:
// 		return true
// 	}
// 	return false
// }

// FetchTags fetches tags for each of files.
// Does not call l.SetTotal(). This should be called by the caller of FetchTags.
func FetchTags(files <-chan common.Record, l FetchLogger) error {
	panic("TODO")
	return nil
	// res := make(chan tagFetchResponse, 4)
	// left := make(chan bool, 4) // Still files left to process
	// go func() {
	// 	for kv := range files {
	// 		left <- true
	// 		fetchTags <- tagFetchRequest{
	// 			KeyValue: kv,
	// 			res:      res, // Simply fan-in the responses
	// 		}
	// 	}
	// 	close(left)
	// }()

	// for range left {
	// 	switch r := <-res; r.err {
	// 	case nil:
	// 		if r.tags != nil {
	// 			if err := writeFetchedTags(r.KeyValue, r.tags); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		l.Done(r.KeyValue)
	// 	case errNoMatch:
	// 		l.NoMatch(r.KeyValue)
	// 	default:
	// 		l.Err(r.err)
	// 	}
	// }

	// l.Close()
	// return nil
}

// func writeFetchedTags(kv KeyValue, tags [][]byte) error {
// 	kv.setFetchedTags()
// 	return writeRecord(kv, tags)
// }

// Fetch tags of a single file
func FetchSingleFileTags(r common.Record) (err error) {
	panic("TODO")
	return nil
	// res := make(chan tagFetchResponse)
	// fetchTags <- tagFetchRequest{
	// 	KeyValue: kv,
	// 	res:      res,
	// }
	// r := <-res
	// switch r.err {
	// case nil:
	// 	err = writeFetchedTags(kv, r.tags)
	// case errNoMatch:
	// 	err = nil
	// }
	// return
}
