package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func get_connection(file_path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", file_path)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func create_tbls(db *sql.DB) error {
	statement := `
		PRAGMA foreign_keys = ON;
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;

		CREATE TABLE IF NOT EXISTS blob (
			id		TEXT PRIMARY KEY,
			ctime 		INTEGER NOT NULL,
			size 		INTEGER NOT NULL,
			data 		BLOB
		);

		CREATE TABLE IF NOT EXISTS manifest (
			id		TEXT PRIMARY KEY,
			ctime		INTEGER NOT NULL,
			user		TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS manifest_file (
			m_id		TEXT NOT NULL REFERENCES manifest(id),
			path		TEXT NOT NULL,
			blob		TEXT NOT NULL REFERENCES blob(id),
			old_path	TEXT,
			perm		INTEGER,
			PRIMARY KEY (m_id, path)
		);

		CREATE TABLE IF NOT EXISTS checkin (
			id		TEXT PRIMARY KEY,
			ctime		INTEGER NOT NULL,
			user		TEXT NOT NULL,
			comment		TEXT,
			m_id		TEXT NOT NULL REFERENCES manifest(id)
		);
		CREATE TABLE IF NOT EXISTS checkin_parents (
			c_id		TEXT NOT NULL REFERENCES checkin(id),
			p_id		TEXT NOT NULL REFERENCES checkin(id),
			is_primary	BOOLEAN DEFAULT TRUE,
			PRIMARY KEY(c_id, p_id)
		);

		CREATE TABLE IF NOT EXISTS ref (
			id		TEXT PRIMARY KEY,
			ctime		INTEGER NOT NULL,
			user		TEXT NOT NULL,
			type		TEXT NOT NULL,
			comment		TEXT
		);

		CREATE TABLE IF NOT EXISTS ref_change (
			ref_id		TEXT NOT NULL REFERENCES ref(id),
			ctime		INTEGER NOT NULL,
			op		TEXT NOT NULL CHECK(op IN ('+','-')),
			user		TEXT NOT NULL,
			key		TEXT,
			value		TEXT,
			target		TEXT,
			comment		TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_ref_id on ref(id);
		CREATE INDEX IF NOT EXISTS idx_checkin_id on checkin(id);
		CREATE INDEX IF NOT EXISTS idx_manifest_id on manifest(id);
		CREATE INDEX IF NOT EXISTS idx_ref_change_ref_id on ref_change(ref_id);
		`
	_, err := db.Exec(statement)
	if err != nil {
		return err
	}
	log.Print("Storage tables created")
	return nil
}

func main() {
	db, err := get_connection("./orc.db")
	if err != nil {
		log.Panicf("Couldn't get connection: %s", err)
	}

	err = create_tbls(db)
	if err != nil {
		log.Panicf("Couldn't create tables: %s", err)
	}

	in := os.Args[1]
	data := []byte(in)
	id, err := put_blob(db, data)
	if err != nil {
		log.Panicf("Couldn't create blob: %s", err)
	}

	b, err := get_blob(db, id)
	if err != nil {
		log.Panicf("Couldn't create blob: %s", err)
	}

	log.Printf("ID=%s\tCTIME=%s\tSIZE=%d\tDATA=%s", b.id, time.Unix(b.ctime, 0).UTC(), b.size, string(b.data))
}
