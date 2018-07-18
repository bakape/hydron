package fetch

import (
	"net/http"
	"strings"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/tags"
	"golang.org/x/net/html"
)

var fetchTags = make(chan request)

// Number of parallel fetch workers
const FetcherCount = 8

type request struct {
	md5 string
	res chan<- response
}

type response struct {
	tags []common.Tag
	err  error
}

// Spawn 8 goroutines to perform parallel fetches
func init() {
	for i := 0; i < FetcherCount; i++ {
		go func() {
			for {
				req := <-fetchTags
				tags, err := fetchFromGelbooru(req.md5)
				req.res <- response{
					tags: tags,
					err:  err,
				}
			}
		}()
	}
}

// Find the tag list in the document using depth-first recursion
func findTags(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			if attr.Key == "id" && attr.Val == "searchTags" {
				return n
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		found := findTags(c)
		if found != nil {
			return found
		}
	}

	return nil
}

func scrapeTags(n *html.Node, t *[]common.Tag) {
	for n := n.FirstChild; n != nil; n = n.NextSibling {
		if n.Type != html.ElementNode || n.Data != "li" {
			continue
		}
		class := ""
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.HasPrefix(attr.Val, "tag-type-") {
				class = attr.Val
				break
			}
		}
		if class == "" {
			continue
		}

		// Get last <a> child
		var lastA *html.Node
		for n := n.FirstChild; n != nil; n = n.NextSibling {
			if n.Type == html.ElementNode && n.Data == "a" {
				lastA = n
			}
		}
		if lastA == nil {
			continue
		}

		text := lastA.FirstChild
		if text == nil || text.Type != html.TextNode || text.Data == "" {
			continue
		}
		tag := tags.Normalize(text.Data, common.Gelbooru)
		switch class {
		case "tag-type-artist":
			tag.Type = common.Author
		case "tag-type-character":
			tag.Type = common.Character
		case "tag-type-copyright":
			tag.Type = common.Series
		default:
			tag.Type = common.Undefined
		}
		*t = append(*t, tag)
	}
}

func fetchFromGelbooru(md5 string) (tags []common.Tag, err error) {
	r, err := http.Get(
		"https://gelbooru.com/index.php?page=post&s=list&tags=md5:" + md5)
	if err != nil {
		return
	}
	defer r.Body.Close()

	doc, err := html.Parse(r.Body)
	if err != nil {
		return
	}
	tagList := findTags(doc)
	if tagList == nil {
		return
	}

	tags = make([]common.Tag, 0, 64)
	scrapeTags(tagList, &tags)

	return
}

// Fetch tags from gelbooru for a given MD5 hash.
// Return tags and last change time.
func FetchTags(md5 string) ([]common.Tag, error) {
	ch := make(chan response)
	fetchTags <- request{md5, ch}
	res := <-ch
	return res.tags, res.err
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
