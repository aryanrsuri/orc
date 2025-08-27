package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"
)

// FIXME: One SQL thread, rollback on any errors
func put_checkin(db *sql.DB, C string, root string, ignore []string) (string, error) {
	L, err := put_manifest(db, root, ignore)
	if err != nil {
		return "", err
	}
	P, err := get_previous_checkin(db)
	if err != nil {
		return "", err
	}

	checkin, err := write_checkin(C, L, P)
	if err != nil {
		return "", err
	}
	id, err := put_blob(db, checkin)
	if err != nil {
		return "", err
	}
	insert_checkin := `INSERT INTO checkin (id, ctime, user, comment, m_id) 
	VALUES (?, ?, ?, ?, ?)`
	insert_checkin_parents := `INSERT INTO checkin_parents (c_id, p_id, is_primary)
	VALUES (?, ?, ?)`
	_, err = db.Exec(insert_checkin, id, time.Now().Unix(), get_user(), C, L)
	if err != nil {
		return "", err
	}
	is_primary := false
	if P == "" {
		P = id
		is_primary = true
	}
	_, err = db.Exec(insert_checkin_parents, id, P, is_primary)
	if err != nil {
		return "", err
	}

	err = os.WriteFile("./.orc/INDEX", []byte(id), os.ModePerm)
	if err != nil {
		return "", err
	}

	return id, nil
}

func get_previous_checkin(db *sql.DB) (string, error) {
	query := "SELECT id from checkin ORDER BY ctime DESC LIMIT 1;"
	var id string
	row, err := db.Query(query)
	if err != nil {
		return "", nil
	}
	row.Next()
	row.Scan(&id)
	return id, nil
}

func write_checkin(C string, L string, P string) ([]byte, error) {
	var checkin string
	if len(P) > 1 {
		checkin = fmt.Sprintf(
			"A C\nC %s\nD %d\nL %s\nP %s\nU %s\n", C, time.Now().Unix(), L, P, get_user())
	} else {
		checkin = fmt.Sprintf(
			"A C\nC %s\nD %d\nL %s\nU %s\n", C, time.Now().Unix(), L, get_user())
	}
	Z, err := hash([]byte(checkin))
	if err != nil {
		return nil, err
	}
	checkin = checkin + fmt.Sprintf(
		"Z %s", Z)

	return []byte(checkin), nil
}
