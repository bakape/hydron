package db

import (
	"database/sql"
	"log"
)

var version = len(migrations)

var migrations = []func(*sql.Tx) error{
	func(tx *sql.Tx) (err error) {
		// Initialize DB
		return execAll(tx,
			`create table main (
				id text not null primary key,
				val text not null
			)`,
			`create table tags (
				id INTEGER PRIMARY KEY,
				type tinyint not null,
				tag text not null
			)`,
			`create unique index i_tag_type on tags(type, tag)`,
			// Enables prefix search optimization for LIKE queries
			`create index i_tags on tags(tag)`,
			`create table images (
				id INTEGER PRIMARY KEY,
				type tinyint not null,
				width int not null,
				height int not null,
				import_time int not null,
				size int not null,
				duration int not null,
				md5 text not null,
				sha1 text not null,
				thumb_width int not null,
				thumb_height int not null,
				thumb_is_png tinyint not null
			)`,
			`create unique index i_image_sha1 on images(sha1)`,
			`create unique index i_image_md5 on images(md5)`,
			`create table image_tags (
				image_id int not null references images on delete cascade,
				tag_id int not null references tags,
				source tinyint not null
			)`,
			`create index i_image_tags on image_tags(image_id)`,
			`create index i_tag_images on image_tags(tag_id)`,
			`insert into main (id, val) values ('version', '1')`,
		)
	},
	func(tx *sql.Tx) (err error) {
		return execAll(tx,
			`create table last_change (
				source tinyint not null,
				image_id int not null references images on delete cascade,
				time bigint not null,
				primary key (source, image_id)
			)`,
			`create index i_last_change_image on last_change(image_id)`,
		)
	},
	func(tx *sql.Tx) (err error) {
		return execAll(tx,
			`create index i_image_size on images(size)`,
			`create index i_image_width on images(width)`,
			`create index i_image_height on images(height)`,
			`create index i_image_duration on images(duration)`,
		)
	},
	func(tx *sql.Tx) (err error) {
		log.Printf(
			"please run `hydron fetch_tags` to fetch scoped tags from Gelbooru")
		return execAll(tx, `drop table last_change`)
	},
}

// Run migrations from version `from`to version `to`
func runMigrations(from, to int) (err error) {
	var tx *sql.Tx
	for i := from; i < to; i++ {
		log.Printf("upgrading database to version %d\n", i+1)
		tx, err = db.Begin()
		if err != nil {
			return
		}

		err = migrations[i](tx)
		if err != nil {
			return rollBack(tx, err)
		}

		// Write new version number
		_, err = tx.Exec(
			`update main set val = $1 where id = 'version'`,
			i+1,
		)
		if err != nil {
			return rollBack(tx, err)
		}

		err = tx.Commit()
		if err != nil {
			return
		}
	}
	return
}

func rollBack(tx *sql.Tx, err error) error {
	if rbErr := tx.Rollback(); rbErr != nil {
		err = WrapError(err.Error(), rbErr)
	}
	return err
}
