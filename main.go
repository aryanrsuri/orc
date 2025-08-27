package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const ROOT string = ".orc"
const DB string = ".orc/project.db"
const INDEX string = ".orc/INDEX"
const HELP string = `
Usage:
  orc <command> [options] [args]

Commands:
  init
  checkin
        -m <message>    Commit message for this checkin
  ref
      Manage references (tags, tickets, etc.).
      Subcommands:
        create       Create a new reference (root). Generates a new ref ID.
        edit <id>    Edit an existing reference by ID. Generates a new delta.
        -m <comment>       Comment or message for this ref or delta (required)
        -t <type>          Type of the ref (default: "default") [only for create]
        -k <key=value>     Add a key/value pair (repeatable)
        -kr <key=value>    Remove a key/value pair (repeatable)
        -l <checkin>       Link to a checkin hash (repeatable)
  status
  help

Examples:
  orc init
  orc checkin -m "Initial commit"
  orc ref create -m "Release candidate" -t tag -k label=rc5 -l 123abc
  orc ref edit 77f5c22da21df75d60932e1cdbbd8a30293bc742 -m "Mark resolved" -k status=closed -kr status=in-review -l 15a1cd3ffd681793503d35946a047d8871e22c81
  orc status
`

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
			fmt.Print("Orc project not found, creating one...\n")
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
			type		INTEGER,
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
			uid		TEXT NOT NULL UNIQUE,
			user		TEXT NOT NULL,
			type		TEXT NOT NULL,
			comment		TEXT
		);

		CREATE TABLE IF NOT EXISTS ref_delta (
			id		TEXT NOT NULL,
			ref_id		TEXT NOT NULL REFERENCES ref(id),
			ctime		INTEGER NOT NULL,
			user		TEXT NOT NULL,
			PRIMARY KEY(id, ref_id)
		);

		CREATE INDEX IF NOT EXISTS idx_ref_id on ref(id);
		CREATE INDEX IF NOT EXISTS idx_checkin_id on checkin(id);
		CREATE INDEX IF NOT EXISTS idx_manifest_id on manifest(id);
		CREATE INDEX IF NOT EXISTS idx_ref_delta_id on ref_delta(id);
		`
	_, err := db.Exec(statement)
	if err != nil {
		return err
	}
	return nil
}

type config struct {
	root   string
	ignore []string
}

type multi_flag []string
func (m *multi_flag) String() string   { return strings.Join(*m, ",") }
func (m *multi_flag) Set(v string) error { *m = append(*m, v); return nil }


func main() {
	var c config
	c.root = "."
	c.ignore = []string{".git", ".orc", ".DS_Store", "orc", "examples"}
	if len(os.Args) < 2 {
		fmt.Println("Expected some command")
		os.Exit(1)
	}

	switch os.Args[1] {
	default:
		fmt.Println("This command you entered is either gibberish, or yet to be implemented.")
		os.Exit(1)
	case "help":
		fmt.Println(HELP)
		os.Exit(0)
	case "init":
		cmd := flag.NewFlagSet("init", flag.ExitOnError)
		_ = cmd.Parse(os.Args[2:])
		_, err := ensure_root(ROOT, INDEX)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "checkin":
		cmd := flag.NewFlagSet("checkin", flag.ExitOnError)
		comment := cmd.String("m", "", "checkin comment")
		_ = cmd.Parse(os.Args[2:])
		db, err := get_connection(DB)
		if err != nil {
			fmt.Printf("No .orc project found, call `orc init` first.")
			os.Exit(1)
		}
		id, err := put_checkin(db, *comment, c.root, c.ignore)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		now := time.Now().UTC().String()
		fmt.Printf("Checkin `%s` created at %s\nCheckin ID: %s\n", *comment, now, id)
	case "ref":
		cmd := flag.NewFlagSet("ref", flag.ExitOnError)
		comment := cmd.String("m", "", "ref comment")
		ref_type := cmd.String("t", "ticket", "ref type")

		var add_k multi_flag
		var del_k multi_flag
		var links multi_flag

		cmd.Var(&add_k, "k","add key=value")
		cmd.Var(&del_k, "kr","delete key=value")
		cmd.Var(&links, "l","link checkin")

		_ = cmd.Parse(os.Args[2:])
		args := cmd.Args()
		if len(args) < 1 {
			fmt.Printf("expected subcommand\n")
			os.Exit(1)
		}


		db, err := get_connection(DB)
		if err != nil {
			fmt.Printf("No .orc project found, call `orc init` first.\n")
			os.Exit(1)
		}

		K := make(map[string]pair)
		for _, k_v := range add_k {
			parts := strings.SplitN(k_v, "=", 2)
			if len(parts) == 2 {
				K[parts[0]] = pair{byte: '+', string: parts[1]}
			}
		}
		for _, k_v := range del_k {
			parts := strings.SplitN(k_v, "=", 2)
			if len(parts) == 2 {
				K[parts[0]] = pair{byte: '-', string: parts[1]}
			}
		}

		switch args[0] {
		default: 
			fmt.Println("unkown subcommand")
			os.Exit(1)
		case "create":
			id, err := put_ref(db, *comment, K, links, *ref_type)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("ref %s created at %s\nref id: %s\n", *comment, time.Now().UTC().String(), id)
		case "edit":
			if len(args) < 2 {
				fmt.Println("no ref id provided")
				os.Exit(1)
			}
			ref_id := args[1]
			id, err := put_delta(db, ref_id, *comment, K, links)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println("ref %s updated at %s\ndelta id: %s\n", ref_id, time.Now().UTC().String(), id)
		}



	case "status":
		cmd := flag.NewFlagSet("status", flag.ExitOnError)
		_ = cmd.Parse(os.Args[2:])
		db, err := get_connection(DB)
		if err != nil {
			fmt.Printf("No .orc project found, call `orc init` first.")
			os.Exit(1)
		}
		files, err := walk_dir(c.root, c.ignore)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		F, err := index_blobs(db, files)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		P, err := get_previous_manifest_id(db)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if P != "" {
			s, err := compare_manifest(db, F, P)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(format_state(s, P))
			os.Exit(0)

		} else {
			var s state
			for file, _ := range F {
				s.U = append(s.U, file)
			}
			fmt.Println(format_state(&s, ""))
			os.Exit(0)
		}

	}
}
