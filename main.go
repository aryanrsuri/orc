package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const ROOT string = ".orc"
const DB string = ".orc/project.db"
const INDEX string = ".orc/INDEX"

func create_root(root string, index string) error {
	err := os.Mkdir(root, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(index, []byte{}, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func ensure_root(root string, index string) (*sql.DB, error) {
	if _, err := os.Stat(".orc"); err != nil {
		if os.IsNotExist(err) {
			log.Print("Orc project not found, creating one")
			err = create_root(root, index)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	db, err := get_connection(DB)
	if err != nil {
		return nil, err
	}
	err = create_tbls(db)
	if err != nil {
		return nil, err
	}

	return db, err
}

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
			blob_id		TEXT NOT NULL REFERENCES blob(id),
			ctime		INTEGER NOT NULL,
			PRIMARY KEY (ref_id, blob_id)
		);

		CREATE INDEX IF NOT EXISTS idx_ref_id on ref(id);
		CREATE INDEX IF NOT EXISTS idx_checkin_id on checkin(id);
		CREATE INDEX IF NOT EXISTS idx_manifest_id on manifest(id);
		CREATE INDEX IF NOT EXISTS idx_ref_change_blob_id on ref_change(blob_id);
		`
	_, err := db.Exec(statement)
	if err != nil {
		return err
	}
	return nil
}

func main() {

	// ORC INIT
	db, err := ensure_root(ROOT, INDEX)
	if err != nil {
		log.Panicf("There was an issue creating the orc project: %s", err)
	}

	// TESTING PUT BLOB
	data, err := read_to_bytes("./examples/ref")
	if err != nil {
		log.Panicf("Couldn't read file: %s", err)
	}
	id, err := put_blob(db, data)
	if err != nil {
		log.Panicf("Couldn't create blob: %s", err)
	}

	// TESTING GET BLOB
	b, err := get_blob(db, id)
	if err != nil {
		log.Panicf("Couldn't create blob: %s", err)
	}

	// TESTING PARSE ARTIFACT
	ref, err := get_artifact(nil, b)

	if err != nil {
		panic(err)
	}
	fmt.Println(ref)
}
