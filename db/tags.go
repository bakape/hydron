package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/bakape/hydron/common"
)

/*
Add tags to an image. All tags must be of same TagSource.
imageID: internal ID of image
tags: tags to add
*/
func AddTags(imageID int64, tags []common.Tag) error {
	return InTransaction(func(tx *sql.Tx) (err error) {
		return AddTagsTx(tx, imageID, tags)
	})
}

/*
Add tags to an image in an exisiting transaction. All tags must be of same
TagSource.
tx: transaction to use
imageID: internal ID of image
tags: tags to add
*/
func AddTagsTx(tx *sql.Tx, imageID int64, tags []common.Tag) (
	err error,
) {
	if driver == "postgres" {
		// Because SERIAL tag IDs can collide
		_, err = tx.Exec("lock table tags")
		if err != nil {
			return
		}
	}
	var tagID int64
	for _, t := range tags {
		err = selectTagID().
			Where("tag = ? and type = ?", t.Tag, int(t.Type)).
			RunWith(tx).
			QueryRow().
			Scan(&tagID)
		switch err {
		case nil:
		case sql.ErrNoRows:
			err = nil
			tagID, err = getLastID(tx, sq.
				Insert("tags").
				Columns("tag", "type").
				Values(t.Tag, t.Type))
			if err != nil {
				return
			}
		default:
			return
		}

		_, err = sq.
			Insert("image_tags").
			Columns("image_id", "tag_id", "source").
			Values(imageID, tagID, t.Source).
			RunWith(tx).
			Exec()
		if err != nil {
			return
		}
	}

	return
}

// Remove specific tags from an image
func RemoveTags(imageID int64, tags []common.Tag) error {
	return InTransaction(func(tx *sql.Tx) (err error) {
		for _, t := range tags {
			_, err = sq.
				Delete("image_tags").
				Where(
					`image_id = ?
						and source = ?
						and tag_id = (
							select id from tags
							where tag = ? and type = ?)`,
					imageID, t.Source, t.Tag, t.Type,
				).
				RunWith(tx).
				Exec()
			if err != nil {
				return
			}
		}
		return
	})
}

// Try to match an incomplete tag against a list of postfixes, return matches
func matchPost(re *regexp.Regexp, prefix string,
	postfixes []string) (matches []string) {
	for _, p := range postfixes {
		matched := re.FindString(p)
		if len(matched) != 0 {
			matches = append(matches, prefix+matched)
		}
	}
	return
}

// Attempt to complete a tag by suggesting up to 20 possible tags for a prefix
func CompleteTag(s string) (tags []string, err error) {
	const maxCap = 20
	tags = make([]string, 0, maxCap)
	typeQ := ""
	prefix := ""
	if s[0] == '-' {
		prefix = "-"
		s = s[1:]
	}

	i := strings.IndexByte(s, ':')
	if i != -1 {
		// Complete postfix
		re := regexp.MustCompile(`^` + regexp.QuoteMeta(s[i+1:]) + `.*`)
		switch s[:i] {
		case "undefined":
			prefix += s[:i+1]
			s = s[i+1:]
			typeQ = fmt.Sprintf(" and type = %d", common.Undefined)
		case "artist":
			fallthrough
		case "author":
			prefix += s[:i+1]
			s = s[i+1:]
			typeQ = fmt.Sprintf(" and type = %d", common.Author)
		case "character":
			prefix += s[:i+1]
			s = s[i+1:]
			typeQ = fmt.Sprintf(" and type = %d", common.Character)
		case "copyright":
			fallthrough
		case "series":
			prefix += s[:i+1]
			s = s[i+1:]
			typeQ = fmt.Sprintf(" and type = %d", common.Series)
		case "meta":
			prefix += s[:i+1]
			s = s[i+1:]
			typeQ = fmt.Sprintf(" and type = %d", common.Meta)
		case "rating":
			tags = matchPost(re, "rating:",
				[]string{"safe", "questionable", "explicit"})
			return
		case "system":
			tags = matchPost(re, "system:",
				[]string{"size", "width", "height", "duration", "tag_count"})
			return
		case "order":
			prefix = s[:i+1]
			if s[i+1:] != "" {
				if s[i+1] == '-' {
					prefix = "order:-"
					re = regexp.MustCompile(`^` + regexp.QuoteMeta(s[i+2:]) + `.*`)
				}
			}
			tags = matchPost(re, prefix,
				[]string{"size", "width", "height", "duration", "tag_count", "random"})
			return
		case "limit":
			return
		default:
			// Continue as regular tag
		}
	} else {
		// Complete prefix
		re := regexp.MustCompile(`^` + regexp.QuoteMeta(s) + `.*`)
		for _, t := range []string{"undefined", "artist", "author", "character",
			"copyright", "series", "meta", "rating", "system", "order", "limit"} {
			matched := re.FindString(t)
			if matched != "" {
				tags = append(tags, prefix+matched+":")
			}
		}
	}

	r, err := sq.Select("tag").
		Distinct().
		From("tags").
		Where("tag like ? || '%' escape '$'"+typeQ, // Escape underscores
			strings.Replace(s, "_", "$_", -1)).
		OrderBy("tag").
		Limit(maxCap).
		Query()
	if err != nil {
		return
	}
	defer r.Close()

	var tag string
	for r.Next() {
		err = r.Scan(&tag)
		if err != nil {
			return
		}
		tags = append(tags, prefix+tag)
		if len(tags) == maxCap {
			break
		}
	}
	err = r.Err()
	return
}

// Update tags for a given image and TagSource.
// All tags must be of same TagSource.
func UpdateTags(imageID int64, tags []common.Tag, source common.TagSource,
) error {
	return InTransaction(func(tx *sql.Tx) (err error) {
		// Remove old tags
		_, err = sq.
			Delete("image_tags").
			Where(squirrel.Eq{
				"image_id": imageID,
				"source":   source,
			}).
			RunWith(tx).
			Exec()
		if err != nil {
			return
		}

		err = AddTagsTx(tx, imageID, tags)
		return
	})
}
