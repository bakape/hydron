package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/bakape/hydron/v3/common"
	"github.com/bakape/hydron/v3/files"
)

type IDAndMD5 struct {
	ID  int64
	MD5 string
}

func selectTagID() squirrel.SelectBuilder {
	return sq.Select("id").From("tags")
}

// Searches images by params and executes function on each.
// Matches all images, if page.Filters is empty.
// The images will only contain the bare minimum data to render thumbnails.
func SearchImages(page *common.Page, paginate bool,
	fn func(common.CompactImage) error,
) (err error) {
	var (
		pos []int64 // Must all match
		neg []int64 // Must not match any
	)

	// Get tag IDs for all tag params
	if len(page.Filters.Tag) != 0 {
		err = InTransaction(func(tx *sql.Tx) (err error) {
			var id int64
			for _, t := range page.Filters.Tag {
				var q squirrel.SelectBuilder
				if t.Type == common.Undefined {
					// Undefined tag type matches the first tag type
					// available. User should be specific about tag types,
					// when matching by artist, character, series, etc.
					q = selectTagID().
						Where("tag = ?", t.Tag).
						OrderBy("type asc").
						Limit(1)
				} else {
					q = selectTagID().
						Where("type = ? and tag = ?", t.Type, t.Tag)
				}
				err = q.RunWith(tx).QueryRow().Scan(&id)
				switch err {
				case nil:
				case sql.ErrNoRows:
					// Missing negations can be ignored
					if t.Negative {
						err = nil
						continue
					}
					// But missing positives would result in matching
					// nothing anyway
					return
				default:
					return
				}
				if t.Negative {
					neg = append(neg, id)
				} else {
					pos = append(pos, id)
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

	// Build queries

	q := sq.Select("sha1", "type", "thumb_width", "thumb_height").
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
	for _, s := range page.Filters.System {
		var p string
		switch s.Type {
		case common.Size:
			p = "size"
		case common.Width:
			p = "width"
		case common.Height:
			p = "height"
		case common.Duration:
			p = "duration"
		case common.TagCount:
			p = `(select count(*)
				from image_tags as it
				where it.image_id = i.id)`
		case common.Type:
			p = "type"
		}
		apply("%s %s %d", p, s.Comparator, s.Value)
	}
	for _, s := range page.Filters.SystemStr {
		var p string
		switch s.Type {
		case common.MD5Field:
			p = "md5"
		case common.SHA1Field:
			p = "sha1"
		case common.Name:
			p = "name"
		}
		// Need to use prepared statements instead of string concat to
		// ensure values are treated as strings
		q = q.Where(p+" = ?", s.Tag)
		count = count.Where(p+" = ?", s.Tag)
	}

	{
		var by string
		mode := "asc"
		if page.Order.Reverse {
			mode = "desc"
		}
		switch page.Order.Type {
		// SQLite does not guarantee any order without an ORDER BY statement
		case common.None:
			by = "i.id"
		case common.BySize:
			by = "size"
		case common.ByWidth:
			by = "width"
		case common.ByHeight:
			by = "height"
		case common.ByDuration:
			by = "duration"
		case common.ByTagCount:
			by = `(select count(*)
				from image_tags as it
				where it.image_id = i.id)`
		case common.Random:
			by = "random()"
		}
		q = q.OrderBy(fmt.Sprintf("%s %s", by, mode))
	}

	if paginate && (page.Limit == 0 || page.Limit > common.PageSize) {
		page.Limit = common.PageSize
	}
	if page.Limit != 0 {
		q = q.Limit(uint64(page.Limit))
	}
	if paginate {
		q = q.Offset(uint64(page.Page * page.Limit))
	}

	// Read all matched rows
	r, err := q.Query()
	if err != nil {
		return
	}
	defer r.Close()
	var rec common.CompactImage
	for r.Next() {
		err = r.Scan(&rec.SHA1, &rec.Type, &rec.Thumb.Width, &rec.Thumb.Height)
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
	var total uint
	err = count.QueryRow().Scan(&total)
	if err != nil {
		return
	}
	page.PageTotal = total / page.Limit
	if total%page.Limit != 0 {
		page.PageTotal++
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
	var srcType common.FileType
	err = InTransaction(func(tx *sql.Tx) (err error) {
		err = sq.
			Select("type").
			From("images").
			Where("sha1 = ?", id).
			RunWith(tx).
			QueryRow().
			Scan(&srcType)
		if err != nil {
			return
		}

		_, err = sq.
			Delete("images").
			Where("sha1 = ?", id).
			RunWith(tx).
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
		files.ThumbPath(id),
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
		err = sq.
			Select(
				"type", "sha1", "thumb_width", "thumb_height",
				"width", "height", "import_time", "size", "duration", "md5",
				"id", "name").
			From("images").
			Where("sha1 = ?", sha1).
			RunWith(tx).
			QueryRow().
			Scan(
				&img.Type, &img.SHA1, &img.Thumb.Width, &img.Thumb.Height,
				&img.Width, &img.Height, &img.ImportTime, &img.Size,
				&img.Duration, &img.MD5, &img.ID, &img.Name,
			)
		if err != nil {
			return
		}

		r, err := sq.Select("source", "type", "tag").
			From("image_tags").
			Join("tags on image_tags.tag_id = tags.id").
			Where("image_id = ?", img.ID).
			RunWith(tx).
			Query()
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

func GetImageIDAndMD5(sha1 string) (pair IDAndMD5, err error) {
	err = sq.Select("id", "md5").
		From("images").
		Where("sha1 = ?", sha1).
		QueryRow().
		Scan(&pair.ID, &pair.MD5)
	return
}

// Get IDs and MD5 hashes of all images that can have tags on danbooru
func GetDanbooruTaggable() (pairs []IDAndMD5, err error) {
	r, err := sq.Select("id", "md5").
		From("images").
		Where(
			"type in (?,?,?,?)",
			common.JPEG, common.PNG, common.GIF, common.WEBM,
		).
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
	inner, args, err := sq.Select("1").
		From("images").
		Where("sha1 = ?", sha1).
		ToSql()
	if err != nil {
		return
	}
	err = db.QueryRow(fmt.Sprintf("select exists (%s)", inner), args...).
		Scan(&imported)
	return
}

// Write image and its tags to database and return the image ID
func WriteImage(i common.Image) (id int64, err error) {
	err = InTransaction(func(tx *sql.Tx) (err error) {
		q := sq.Insert("images").
			Columns(
				"type", "width", "height", "import_time", "size", "duration",
				"md5", "sha1", "thumb_width", "thumb_height", "name",
			).
			Values(
				i.Type, i.Width, i.Height, i.ImportTime, i.Size, i.Duration,
				i.MD5, i.SHA1, i.Thumb.Width, i.Thumb.Height, i.Name,
			)
		id, err = getLastID(tx, q)
		if err != nil {
			return
		}

		err = AddTagsTx(tx, id, i.Tags)
		return
	})
	return
}

// SetName sets an image's name
func SetName(id int64, name string) error {
	_, err := sq.Update("images").
		Set("name", name).
		Where("id = ?", id).
		Exec()

	return err
}
