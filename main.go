package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

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
			fmt.Print("Orc project not found, creating one...")
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
	if err := db.Ping(); err != nil {
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

type config struct {
	root string
	ignore []string
}

func main() {
	var c config
	c.root = "."
	c.ignore = []string{".git", ".orc", ".DS_Store", "orc"}
	if len(os.Args) < 2 {
		fmt.Println("Expected some command")
		os.Exit(1)
	}

	switch os.Args[1] {
	default:
		fmt.Println("orc is a simple scm\n\nCommands\n------------\nhelp: this message\ninit: create a new .orc project\ncheckin: checkin source code\nlog: review your checkins")
		os.Exit(1)
	case "help":
		fmt.Println("orc is a simple scm\n\nCommands\n------------\nhelp: this message\ninit: create a new .orc project\ncheckin: checkin source code\nlog: review your checkins")
		os.Exit(1)
	case "init":
		cmd := flag.NewFlagSet("init", flag.ExitOnError)
		_ = cmd.Parse(os.Args[2:])
		_, err := ensure_root(ROOT, INDEX)
		if err != nil {
			fmt.Printf("Sorry, can't do that mate: %s\n", err)
			os.Exit(1)
		}
	case "checkin":
		cmd := flag.NewFlagSet("checkin", flag.ExitOnError)
		message := cmd.String("m", "", "commit message")
		_ = cmd.Parse(os.Args[2:])
		db, err := get_connection(DB)
		if err != nil {
			fmt.Printf("No .orc project found, call ``init`` first.")
			os.Exit(1)
		}
		id, err := put_checkin(db, *message, c.root, c.ignore)
		if err != nil {
			fmt.Printf("Sorry, can't do that mate: %s\n", err)
			os.Exit(1)
		}
		now := time.Now().UTC().String()
		fmt.Printf("Checkin created at %s\nCheckin ID: %s", now, id)
	}
}
