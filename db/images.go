package db

import (
	"database/sql"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/files"
	"github.com/bakape/hydron/tags"
)

type IDAndMD5 struct {
	ID  int64
	MD5 string
}

// Searches images by params and executes function on each.
// Matches all images, if params = "".
// The returned images will only contain the bare minimum data to render
// thumbnails.
// TODO: System tags (requires more indexes)
// TODO: Sorting (requires more indexes)
func SearchImages(params string, fn func(common.CompactImage) error) (
	err error,
) {
	// Get tag IDs for all params
	var (
		split = strings.Split(params, " ")
		pos   []uint64
		neg   []uint64
	)
	if len(split) != 0 {
		err = InTransaction(func(tx *sql.Tx) (err error) {
			qWithType := lazyPrepare(tx,
				`select id from tags
				where type = ? and tag = ?`)
			qAnyType := lazyPrepare(tx, `select id from tags where tag = ?`)
			var id uint64
			for _, s := range split {
				s = strings.TrimSpace(s)
				if len(s) == 0 {
					continue
				}
				isNeg := false
				if s[0] == '-' {
					isNeg = true
					s = s[1:]
				}
				t := tags.Normalize(s, common.User)

				// Undefined tag type matches all
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
					// But missing positives would result in mathcing nothing
					// anyway
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
		}
	}

	// Build query
	q := sq.Select(
		"sha1", "type", "thumb_is_png", "thumb_width", "thumb_height",
	).
		From("images as i")
	for _, id := range pos {
		q = q.Where(
			`exists (
				select 1
				from image_tags as it
				where it.image_id = i.id and it.tag_id = ?
			)`,
			id,
		)
	}
	if len(neg) != 0 {
		b := make([]byte, 1, 128)
		b[0] = '('
		for i, id := range neg {
			if i != 0 {
				b = append(b, ',')
			}
			b = strconv.AppendUint(b, id, 10)
		}
		b = append(b, ')')
		q = q.Where(
			`not exists (
				select 1
				from image_tags as it
				where it.image_id = i.id and it.tag_id in ?
			)`,
			string(b),
		)
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
	return r.Err()
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

// Prepare statement for fetching internal image ID by SHA1 for transaction
func getImageID(tx *sql.Tx) preparedStatement {
	return lazyPrepare(tx, `select id from images where sha1 = ?`)
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
