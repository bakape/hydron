package db

import (
	"database/sql"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/bakape/hydron/common"
)

/*
Add tags to an image. All tags must be of same TagSource.
imageID: internal ID of image
tags: tags to add
lastChange: last tag modification time to write
*/
func AddTags(imageID int64, tags []common.Tag, lastChange time.Time) error {
	return InTransaction(func(tx *sql.Tx) (err error) {
		return AddTagsTx(tx, imageID, tags, lastChange)
	})
}

/*
Add tags to an image in an exisiting transaction. All tags must be of same
TagSource.
tx: transaction to use
imageID: internal ID of image
tags: tags to add
lastChange: last tag modification time to write
*/
func AddTagsTx(tx *sql.Tx, imageID int64, tags []common.Tag,
	lastChange time.Time) (
	err error,
) {
	getID := lazyPrepare(tx, `select id from tags where tag = ? and type = ?`)
	insertTag := lazyPrepare(tx, `insert into tags (tag, type) values (?, ?)`)
	insertRelation := lazyPrepare(tx,
		`insert or ignore into image_tags (image_id, tag_id, source)
		values (?, ?, ?)`)
	writeLastChange := lazyPrepare(tx,
		`insert or replace into last_change (source, image_id, time)
		values (?, ?, ?)`)
	var (
		tagID    int64
		res      sql.Result
		affected = make(map[common.TagSource]bool, 4)
	)
	for _, t := range tags {
		err = getID.QueryRow(t.Tag, t.Type).Scan(&tagID)
		switch err {
		case nil:
		case sql.ErrNoRows:
			err = nil
			res, err = insertTag.Exec(t.Tag, t.Type)
			if err != nil {
				return
			}
			tagID, err = res.LastInsertId()
			if err != nil {
				return
			}
		default:
			return
		}

		_, err = insertRelation.Exec(imageID, tagID, t.Source)
		if err != nil {
			return
		}

		affected[t.Source] = true
	}

	// Write last change timestamps
	unix := lastChange.Unix()
	for s := range affected {
		_, err = writeLastChange.Exec(s, imageID, unix)
		if err != nil {
			return
		}
	}

	return
}

// Remove specific tags from an image
func RemoveTags(imageID int64, tags []common.Tag) error {
	return InTransaction(func(tx *sql.Tx) (err error) {
		q := lazyPrepare(tx,
			`delete from image_tags
			where image_id = ?
				and source = ?
				and tag_id = (
					select id from tags
					where tag = ? and type = ?)`)
		for _, t := range tags {
			_, err = q.Exec(imageID, t.Source, t.Tag, t.Type)
			if err != nil {
				return
			}
		}
		return
	})
}

// Attempt to complete a tag by suggesting up to 10 possible tags for a prefix
func CompleTag(s string) (tags []string, err error) {
	r, err := sq.Select("tag").
		Distinct().
		From("tags").
		Where("tag like ? || '%' escape '$'", // Escape underscores
			strings.Replace(s, "_", "$_", -1)).
		OrderBy("tag").
		Limit(10).
		Query()
	if err != nil {
		return
	}
	defer r.Close()

	tags = make([]string, 0, 10)
	var tag string
	for r.Next() {
		err = r.Scan(&tag)
		if err != nil {
			return
		}
		tags = append(tags, tag)
	}
	err = r.Err()
	return
}

// Get time of last change for fetched tag resources
func getLastChange(tx *sql.Tx, imageID int64, source common.TagSource) (
	time.Time, error,
) {
	var t uint64
	err := withTransaction(tx, sq.
		Select("time").
		From("last_change").
		Where(squirrel.Eq{
			"source":   source,
			"image_id": imageID,
		}),
	).
		QueryRow().
		Scan(&t)
	if err == sql.ErrNoRows {
		err = nil
	}
	return time.Unix(int64(t), 0), err
}

// Update tags for a given image and TagSource.
// All tags must be of same TagSource.
// Tags are only updated, if the change timestamp is newer than the last.
func UpdateTags(imageID int64, tags []common.Tag, source common.TagSource,
	timeStamp time.Time,
) error {
	return InTransaction(func(tx *sql.Tx) (err error) {
		lastChange, err := getLastChange(tx, imageID, source)
		if err != nil || !lastChange.Before(timeStamp) {
			return
		}

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

		err = AddTagsTx(tx, imageID, tags, timeStamp)
		return
	})
}
