package db

import (
	"database/sql"

	"github.com/bakape/hydron/common"
)

// Add tags to an image
func AddTags(imageID uint64, tags []common.Tag) error {
	return InTransaction(func(tx *sql.Tx) (err error) {
		getID, insertTag, insertRelation := prepareForTagInsertion(tx)
		var (
			tagID int64
			res   sql.Result
		)
		for _, t := range tags {
			err = getID.QueryRow(t.Tag, t.Type).Scan(&tagID)
			switch err {
			case nil:
			case sql.ErrNoRows:
				err = nil
				res, err = insertTag.Exec(t.Tag, t.Source)
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

			_, err = insertRelation.Exec(imageID, tagID)
			if err != nil {
				return
			}
		}
		return
	})
}

// Prepare tag insertion queries for transaction
func prepareForTagInsertion(tx *sql.Tx) (
	getID, insertTag, insertRelation preparedStatement,
) {
	return lazyPrepare(tx, `select id from tags where tag = ? and type = ?)`),
		lazyPrepare(tx, `insert into tags (tag, source) values (?, ?)`),
		lazyPrepare(tx, `
			insert into image_tags (image_id, tag_id, source)
			values (?, ?, ?)
			on conflict ignore`)
}

func RemoveTags(imageID uint64, tags []common.Tag) error {
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
		Where("tag LIKE ?*", s).
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
