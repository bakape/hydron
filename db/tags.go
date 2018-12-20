package db

import (
	"database/sql"
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
		err = withTransaction(tx, selectTagID().
			Where("tag = ? and type = ?", t.Tag, int(t.Type))).
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

		_, err = withTransaction(tx, sq.
			Insert("image_tags").
			Columns("image_id", "tag_id", "source").
			Values(imageID, tagID, t.Source)).
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
			_, err = withTransaction(tx, sq.
				Delete("image_tags").
				Where(
					`image_id = ?
					and source = ?
					and tag_id = (
						select id from tags
						where tag = ? and type = ?)`,
					imageID, t.Source, t.Tag, t.Type,
				)).
				Exec()
			if err != nil {
				return
			}
		}
		return
	})
}

// Attempt to complete a tag by suggesting up to 10 possible tags for a prefix
func CompleteTag(s string) (tags []string, err error) {
	const maxCap = 20
	tags = make([]string, 0, maxCap)

	// TODO: Inject system tags into array

	r, err := sq.Select("tag").
		Distinct().
		From("tags").
		Where("tag like ? || '%' escape '$'", // Escape underscores
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
		tags = append(tags, tag)
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
		_, err = withTransaction(tx, sq.
			Delete("image_tags").
			Where(squirrel.Eq{
				"image_id": imageID,
				"source":   source,
			}),
		).
			Exec()
		if err != nil {
			return
		}

		err = AddTagsTx(tx, imageID, tags)
		return
	})
}
