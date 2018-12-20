package fetch

import (
	"github.com/bakape/boorufetch"
	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/tags"
)

// Fetch tags from gelbooru for a given MD5 hash.
// Return tags and last change time.
func FetchTags(md5 string) (out []common.Tag, err error) {
	hash, err := boorufetch.DecodeMD5(md5)
	if err != nil {
		return
	}
	post, err := boorufetch.DanbooruByMD5(hash)
	if err != nil || post == nil {
		return
	}
	src, err := post.Tags()
	if err != nil {
		return
	}
	rating, err := post.Rating()
	if err != nil {
		return
	}

	out = make([]common.Tag, len(src)+1)
	for i, t := range src {
		// Hydron and boorufetch tag enums don't match up completely
		out[i] = tags.Normalize(t.Tag, common.Gelbooru)
		if t.Type >= boorufetch.Undefined && t.Type <= boorufetch.Series {
			out[i].Type = common.TagType(t.Type)
		} else if t.Type == boorufetch.Meta {
			out[i].Type = common.Meta
		}
	}
	out[len(out)-1] = common.Tag{
		TagBase: common.TagBase{
			Type: common.Rating,
			Tag:  rating.String(),
		},
		Source: common.Danbooru,
	}
	return
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
