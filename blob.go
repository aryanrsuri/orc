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

func get_blob(db *sql.DB, id string) (*blob, error) {
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

	return &b, nil
}
