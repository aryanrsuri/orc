package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

type blob struct {
	id    string
	ctime int64
	size  int64
	data  []byte
}

type c_artifact struct {
	A string
	C string
	D int64
	F map[string]string
	I string
	fields []map[string][]string
}

func hash(data []byte) (string, error) {
	h := sha256.New()
	_, err := h.Write(data)

	if err != nil {
		return "", err
	}
	id := hex.EncodeToString(h.Sum(nil))

	return id, nil
}

func put_blob(db *sql.DB, data []byte) (string, error) {

	id, err := hash(data)

	if err != nil {
		return "", err
	}

	_, err = db.Exec("INSERT INTO blob (id, ctime, size, data) VALUES (?, ?, ?, ?);",
		id, time.Now().Unix(), len(data), data)

	if err != nil {
		if err.(sqlite3.Error).ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
			log.Printf("Data already exists with id: %s", id)
			return id, nil
		}
		return "", err
	}
	log.Printf("Data inserted with id: %s", id)
	return id, nil
}

func get_blob(db *sql.DB, id string) (blob, error) {
	row, err := db.Query("SELECT * FROM blob WHERE id = ? LIMIT 1;", id)

	if err != nil {
		return nil, err
	}

	var b blob

	row.Next()
	defer row.Close()

	err = row.Scan(&b.id, &b.ctime, &b.size, &b.data)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func parse_blob(b blob) (c_artifact, error) {
	var control c_artifact 
	content := string(b.data)
	lines := strings.Split(content, "\n")
	for n, line := range(lines[:len(lines)-1]) {
		c, field, err := parse_card(line)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			control.c_type = c
		}

		control.fields = append(control.fields, map[string][]string{c: field})
	}

	return control, nil
}


// FIXME: We need to have card specific logic? 
// some way to pack cards into a struct, since,
// for example, there may be more than K cards in a ref
// so a single map won't suffice
func parse_card(line string) (string, []string, error) {
	parts := strings.Fields(line)
	
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("empty line")
	}
	if len(parts) == 1 {
		return "", nil, fmt.Errorf("empty card")
	}
	card := parts[0]
	fields := parts[1:]

	return card, fields, nil
}


