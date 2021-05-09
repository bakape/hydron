package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/bakape/hydron/common"
)

// A little different from tags/filter.go
var systemRegex = regexp.MustCompile(`^([\w_]+)((=|>|>=|<|<=)(\d+)?)?$`)

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
func matchPost(
	s string, pre string, i int, postfixes []string,
) (
	matches []string, err error,
) {
	re, err := regexp.Compile(`^` + regexp.QuoteMeta(s[i:]) + `.*`)
	if err != nil {
		return
	}
	for _, p := range postfixes {
		matched := re.FindString(p)
		if len(matched) != 0 {
			matches = append(matches, pre+s[:i]+matched)
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
		i++
		completeWithPrefix := func(t common.TagType) {
			prefix += s[:i]
			s = s[i:]
			typeQ = fmt.Sprintf(" and type = %d", t)
		}

		switch s[:i-1] {
		case "undefined":
			completeWithPrefix(common.Undefined)
		case "artist", "author":
			completeWithPrefix(common.Author)
		case "character":
			completeWithPrefix(common.Character)
		case "copyright", "series":
			completeWithPrefix(common.Series)
		case "meta":
			completeWithPrefix(common.Meta)
		case "rating":
			tags, err = matchPost(s, prefix, i,
				[]string{"safe", "questionable", "explicit"})
			return
		case "system":
			systemTags := []string{
				"size", "width", "height", "duration", "tag_count",
			}
			substr := s[i:]
			m := systemRegex.FindStringSubmatch(s[i:])
			if m != nil {
				for _, t := range systemTags {
					if m[1] == t {
						// If we have a valid tag but nothing to autocomplete,
						// still return it to show it's valid
						tags = []string{prefix + s}
						return
					}
				}
				substr = m[1]
			}
			tags, err = matchPost(substr, prefix+s[:i], 0, systemTags)
			return
		case "order":
			if prefix != "" {
				// This tag category doesn't work with the "-" prefix
				return
			}
			if s[i:] != "" && s[i] == '-' {
				// Slice after order:-
				i++
			}
			tags, err = matchPost(s, prefix, i, []string{
				"size", "width", "height", "duration",
				"tag_count", "random"})
			return
		case "limit":
			if prefix != "" {
				// This tag category doesn't work with the "-" prefix
				return
			}
			// Remove the autocomplete result if invalid limit
			_, err = strconv.ParseUint(s[i:], 10, 64)
			if err != nil {
				err = nil
				if s[i:] != "" {
					return
				}
			}
			tags = []string{s}
			return
		default:
			// Continue as regular tag
		}
	} else {
		// Complete prefix
		categories := []string{
			"undefined", "artist", "author", "character",
			"copyright", "series", "meta", "rating", "system",
		}
		// These categories don't work with a prefix of "-"
		if prefix == "" {
			categories = append(categories, "order", "limit")
		}

		var re *regexp.Regexp
		re, err = regexp.Compile(`^` + regexp.QuoteMeta(s) + `.*`)
		if err != nil {
			return
		}
		for _, t := range categories {
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
