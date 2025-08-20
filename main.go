package main

import (
	"database/sql"
	"log"
	_ "github.com/mattn/go-sqlite3"
)


func get_connection (file_path string) (*sql.DB, error) {
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
			id		TEXT PRIMARY_KEY,
			size 		INTEGER NOT NULL,
			ctime 		INTEGER NOT NULL,
			data 		BLOB
		);

		CREATE TABLE IF NOT EXISTS manifest (
			id		TEXT PRIMARY_KEY,
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
			message		TEXT,
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
			pairs		JSON, 
			message		TEXT
		);

		CREATE TABLE IF NOT EXISTS ref_change (
			id		TEXT PRIMARY KEY,
			ctime		INTEGER NOT NULL,
			user		TEXT NOT NULL,
			target		TEXT NOT NULL REFERENCES ref(id),
			pair		JSON,
			message		TEXT
		);
		CREATE TABLE IF NOT EXISTS ref_link (
			r_id		TEXT NOT NULL REFERENCES ref(id),
			c_id		TEXT NOT NULL REFERENCES checkin(id),
			type		TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_manifest_id on manifest(id);
		CREATE INDEX IF NOT EXISTS idx_checkin_id on checkin(id);
		CREATE INDEX IF NOT EXISTS idx_ref_id on ref(id);
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
		log.Panicf("Oh, no: %s", err)
	}

	err = create_tbls(db)
	if err != nil {
		log.Panicf("Oh, no: %s", err)
	}
}
