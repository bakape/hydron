package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/files"
	"github.com/bakape/hydron/tags"
)

// Size of page for search query pagination
const pageSize = 100

type IDAndMD5 struct {
	ID  int64
	MD5 string
}

// Searches images by params and executes function on each.
// Matches all images, if params = "".
// The images will only contain the bare minimum data to render thumbnails.
// Returns number of total pages and any error.
// TODO: This pagination will get really slow over time. Some kind of cache?
func SearchImages(params string, page int, fn func(common.CompactImage) error) (
	pages int, err error,
) {
	var (
		split  = strings.Split(params, " ")
		pos    []int64 // Must all match
		neg    []int64 // Must not match any
		system []tags.SystemTag
		order  tags.Ordering
		limit  uint64
	)
	if len(split) != 0 {
		system, order, limit, err = tags.ExtractInternal(&split)
		if err != nil {
			return
		}

		// Get tag IDs for all tag params
		if len(split) != 0 {
			err = InTransaction(func(tx *sql.Tx) (err error) {
				qWithType := lazyPrepare(tx,
					`select id from tags
					where type = ? and tag = ?`)
				// Undefined tag type matches the first tag type available.
				// User should be specific about tag types, when matching by
				// artist, character, series, etc.
				qAnyType := lazyPrepare(tx,
					`select id from tags
					where tag = ?
					order by type asc
					limit 1`)

				var id int64
				for _, s := range split {
					isNeg := false
					if s[0] == '-' {
						isNeg = true
						s = s[1:]
					}
					t := tags.Normalize(s, common.User)

					var rs rowScanner
					if t.Type == common.Undefined {
						rs = qAnyType.QueryRow(t.Tag)
					} else {
						rs = qWithType.QueryRow(t.Type, t.Tag)
					}
					err = rs.Scan(&id)
					switch err {
					case nil:
					case sql.ErrNoRows:
						// Missing negations can be ignored
						if isNeg {
							err = nil
							continue
						}
						// But missing positives would result in matching
						// nothing anyway
						return
					default:
						return
					}
					if !isNeg {
						pos = append(pos, id)
					} else {
						neg = append(neg, id)
					}
				}
				return
			})
			switch err {
			case nil:
			case sql.ErrNoRows:
				err = nil
				return
			default:
				return
			}
		}
	}

	// Build queries

	q := sq.Select("sha1", "type", "thumb_is_png", "thumb_width",
		"thumb_height",
	).
		From("images as  i")
	count := sq.Select("count(*)").From("images as  i")

	// Apply where modification to both the select and count query
	apply := func(format string, args ...interface{}) {
		s := fmt.Sprintf(format, args...)
		q = q.Where(s)
		count = count.Where(s)
	}

	if len(pos) != 0 {
		apply(
			`id in (
				select image_id from image_tags
				where tag_id in %s
				group by image_id
				having count(*) = %d)`,
			formatSet(pos), len(pos),
		)
	}
	if len(neg) != 0 {
		apply(
			`not exists (
				select 1
				from image_tags
				where image_id = id and tag_id in %s)`,
			formatSet(neg),
		)
	}
	for _, s := range system {
		var p string
		switch s.Type {
		case tags.Size:
			p = "size"
		case tags.Width:
			p = "width"
		case tags.Height:
			p = "height"
		case tags.Duration:
			p = "duration"
		case tags.TagCount:
			p = `(select count(*)
				from image_tags as it
				where it.image_id = i.id)`
		}
		apply("%s %s %d", p, s.Comparator, s.Value)
	}

	{
		var by string
		mode := "asc"
		if order.Reverse {
			mode = "desc"
		}
		switch order.Type {
		// SQLite does not guarantee any order without an ORDER BY statement
		case tags.None:
			by = "i.id"
		case tags.BySize:
			by = "size"
		case tags.ByWidth:
			by = "width"
		case tags.ByHeight:
			by = "height"
		case tags.ByDuration:
			by = "duration"
		case tags.ByTagCount:
			by = `(select count(*)
				from image_tags as it
				where it.image_id = i.id)`
		case tags.Random:
			by = "random()"
		}
		q = q.OrderBy(fmt.Sprintf("%s %s", by, mode))
	}

	if limit == 0 || limit > pageSize {
		limit = pageSize
		q = q.Limit(limit).Offset(uint64(page * pageSize))
	} else {
		q = q.Limit(limit)
	}

	// Read all matched rows
	r, err := q.Query()
	if err != nil {
		return
	}
	defer r.Close()
	var rec common.CompactImage
	for r.Next() {
		err = r.Scan(&rec.SHA1, &rec.Type, &rec.Thumb.IsPNG, &rec.Thumb.Width,
			&rec.Thumb.Height)
		if err != nil {
			return
		}
		err = fn(rec)
		if err != nil {
			return
		}
	}
	err = r.Err()
	if err != nil {
		return
	}

	// Read total match count
	var total int
	err = count.QueryRow().Scan(&total)
	if err != nil {
		return
	}
	pages = total / pageSize
	if total%pageSize != 0 {
		pages++
	}

	return
}

// Format SQL set
func formatSet(arr []int64) string {
	b := make([]byte, 1, 256)
	b[0] = '('
	for i, n := range arr {
		if i != 0 {
			b = append(b, ',')
		}
		b = strconv.AppendInt(b, n, 10)
	}
	b = append(b, ')')
	return string(b)
}

// Remove an image from the database by ID. Non-existant files are ignored.
func RemoveImage(id string) (err error) {
	var (
		srcType  common.FileType
		pngThumb bool
	)
	err = InTransaction(func(tx *sql.Tx) (err error) {
		err = withTransaction(tx, sq.
			Select("type", "thumb_is_png").
			From("images").
			Where("sha1 = ?", id),
		).
			QueryRow().
			Scan(&srcType, &pngThumb)
		if err != nil {
			return
		}

		_, err = withTransaction(tx, sq.
			Delete("images").
			Where("sha1 = ?", id),
		).
			Exec()
		return
	})
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil
	default:
		return
	}

	// Remove files
	for _, p := range [...]string{
		files.SourcePath(id, srcType),
		files.ThumbPath(id, pngThumb),
	} {
		err = os.Remove(p)
		switch {
		case err == nil:
		case os.IsNotExist(err):
			err = nil
		default:
			return
		}
	}

	return
}

// Retrieve an image and all it's tags by SHA1 hash
func GetImage(sha1 string) (img common.Image, err error) {
	err = InTransaction(func(tx *sql.Tx) (err error) {
		q := sq.Select(
			"type", "sha1", "thumb_is_png", "thumb_width", "thumb_height",
			"width", "height", "import_time", "size", "duration", "md5", "id",
		).
			From("images").
			Where("sha1 = ?", sha1)
		err = withTransaction(tx, q).QueryRow().Scan(
			&img.Type, &img.SHA1, &img.Thumb.IsPNG, &img.Thumb.Width,
			&img.Thumb.Height,
			&img.Width, &img.Height, &img.ImportTime, &img.Size,
			&img.Duration, &img.MD5, &img.ID,
		)
		if err != nil {
			return
		}

		q = sq.Select("source", "type", "tag").
			From("image_tags").
			Join("tags on image_tags.tag_id = tags.id").
			Where("image_id = ?", img.ID)
		r, err := withTransaction(tx, q).Query()
		if err != nil {
			return
		}
		defer r.Close()
		var tag common.Tag
		img.Tags = make([]common.Tag, 0, 64)
		for r.Next() {
			err = r.Scan(&tag.Source, &tag.Type, &tag.Tag)
			if err != nil {
				return
			}
			img.Tags = append(img.Tags, tag)
		}

		return
	})
	return
}

func GetImageID(sha1 string) (id int64, err error) {
	err = sq.Select("id").
		From("images").
		Where("sha1 = ?", sha1).
		QueryRow().
		Scan(&id)
	return
}

// Get IDs and MD5 hashes of all images that can have tags on gelbooru
func GetGelbooruTaggable() (pairs []IDAndMD5, err error) {
	r, err := sq.Select("id", "md5").
		From("images").
		Where("type in (?,?,?,?)",
			common.JPEG, common.PNG, common.GIF, common.WEBM).
		Query()
	if err != nil {
		return
	}
	defer r.Close()

	pairs = make([]IDAndMD5, 0, 1<<15)
	var pair IDAndMD5
	for r.Next() {
		err = r.Scan(&pair.ID, &pair.MD5)
		if err != nil {
			return
		}
		pairs = append(pairs, pair)
	}
	err = r.Err()
	return
}

// Return, if images is already imported into the database
func IsImported(sha1 string) (imported bool, err error) {
	err = db.QueryRow(
		`select exists (select 1 from images where sha1 = ?)`,
		sha1,
	).
		Scan(&imported)
	return
}

// Write image and its tags to database and return the image ID
func WriteImage(i common.Image) (id int64, err error) {
	err = InTransaction(func(tx *sql.Tx) (err error) {
		q := sq.Insert("images").
			Columns(
				"type", "width", "height", "import_time", "size", "duration",
				"md5", "sha1", "thumb_width", "thumb_height", "thumb_is_png",
			).
			Values(
				i.Type, i.Width, i.Height, i.ImportTime, i.Size, i.Duration,
				i.MD5, i.SHA1, i.Thumb.Width, i.Thumb.Height, i.Thumb.IsPNG,
			)
		var res sql.Result
		res, err = withTransaction(tx, q).Exec()
		if err != nil {
			return
		}
		id, err = res.LastInsertId()
		if err != nil {
			return
		}

		err = AddTagsTx(tx, id, i.Tags, time.Now())
		return
	})
	return
}
