package db

import (
	"database/sql"
	"fmt"
	"log"
)

var version = len(migrations)

// Different keyword for different RDBMS
var autoIncrement string

var migrations = []func(*sql.Tx) error{
	func(tx *sql.Tx) (err error) {
		// Initialize DB
		return execAll(tx,
			`create table main (
				id text not null primary key,
				val text not null
			)`,
			`create table tags (
				id `+autoIncrement+`,
				type smallint not null,
				tag text not null
			)`,
			`create unique index i_tag_type on tags(type, tag)`,
			// Enables prefix search optimization for LIKE queries
			`create index i_tags on tags(tag)`,
			`create table images (
				id `+autoIncrement+`,
				type smallint not null,
				width int not null,
				height int not null,
				import_time int not null,
				size int not null,
				duration int not null,
				md5 text not null,
				sha1 text not null,
				thumb_width int not null,
				thumb_height int not null,
				thumb_is_png boolean not null
			)`,
			`create unique index i_image_sha1 on images(sha1)`,
			`create unique index i_image_md5 on images(md5)`,
			`create table image_tags (
				image_id int not null references images on delete cascade,
				tag_id int not null references tags,
				source smallint not null
			)`,
			`create index i_image_tags on image_tags(image_id)`,
			`create index i_tag_images on image_tags(tag_id)`,
			`insert into main (id, val) values ('version', '1')`,
		)
	},
	func(tx *sql.Tx) (err error) {
		return execAll(tx,
			`create table last_change (
				source smallint not null,
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
		return execAll(tx, `drop table last_change`)
	},
	func(tx *sql.Tx) (err error) {
		return execAll(tx,
			`alter table images add column name text not null default ''`,
		)
	},
	func(tx *sql.Tx) (err error) {
		// No drop column support lol
		if driver != "sqlite3" {
			_, err = tx.Exec(`alter table images drop column thumb_is_png`)
		}
		return
	},
}

// Run migrations from version `from`to version `to`
func runMigrations(from, to int) (err error) {
	switch driver {
	case "postgres":
		autoIncrement = "serial primary key"
	case "mysql":
		autoIncrement = "integer primary key auto_increment"
	case "sqlite3":
		autoIncrement = "integer primary key autoincrement"
	}

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
		_, err = sq.Update("main").
			Set("val", i+1).
			Where("id = 'version'").
			RunWith(tx).
			Exec()
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
		err = fmt.Errorf("%s: %s", err.Error(), rbErr)
	}
	return err
}
