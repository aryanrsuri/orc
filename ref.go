package main

import (
	"database/sql"
	"fmt"
	"time"
)

// FIXME: One SQL thread, rollback on any errors
func put_ref(db *sql.DB, C string,  K map[string]pair, L []string, T string) (string, error) {
	I := uuid()
	ref, err := write_ref(C, I, K, L, T)
	if err != nil {
		return "", err
	}
	id, err := put_blob(db, ref)
	if err != nil {
		return "", err
	}
	insert_ref := `INSERT INTO ref (id, ctime, uid, user, type, comment)
	VALUES (?, ?, ?, ?, ?, ?);`
	insert_ref_delta := `INSERT INTO ref_delta (id, ref_id, ctime, user)
	VALUES (?, ?, ?, ?)`
	_, err = db.Exec(insert_ref, id, time.Now().Unix(),I, get_user(), T, C)
	if err != nil {
		return "", err
	}
	_, err = db.Exec(insert_ref_delta, id, id, time.Now().Unix(), get_user())
	if err != nil {
		return "", err
	}

	return I, nil
}

func write_ref(C string, I string, K map[string]pair, L []string, T string) ([]byte, error) {
	var ref string
	ref = fmt.Sprintf("A R\nC %s\nD %d\nI %s\n", C, time.Now().Unix(), I)
	if len(K) > 0 {
		for tag, change := range K {
			ref = ref + fmt.Sprintf("K %s %s %s\n", string(change.byte), tag, change.string)
		}
	}
	if len(L) > 0 {
		for _, l := range L {
			ref = ref + fmt.Sprintf("L %s \n", l)
		}
	}
	ref = ref + fmt.Sprintf("T %s\nU %s\n", T, get_user())
	Z, err := hash([]byte(ref))
	if err != nil {
		return nil, err
	}
	ref = ref + fmt.Sprintf(
		"Z %s", Z)

	return []byte(ref), nil
}


func put_delta(db *sql.DB, I string, C string, K map[string]pair, L []string) (string, error) {
	P, err := get_previous_delta(db, I)
	if err != nil {
		return "", err
	}
	delta, err := write_delta(C, K, L, P)
	if err != nil {
		return "", err
	}
	id, err := put_blob(db, delta)
	if err != nil {
		return "", err
	}
	ref_id, err := get_root_ref(db, I)
	if err != nil {
		return "", err
	}
	insert_ref_delta := `INSERT INTO ref_delta (id, ref_id, ctime, user)
	VALUES (?, ?, ?, ?)`
	_, err = db.Exec(insert_ref_delta, id, ref_id, time.Now().Unix(), get_user())
	if err != nil {
		return "", err
	}

	return id, nil
}

func write_delta(C string, K map[string]pair, L []string, P string) ([]byte, error) {
	var delta string
	delta = fmt.Sprintf("A D\nC %s\nD %d\n", C, time.Now().Unix())
	if len(K) > 0 {
		for tag, change := range K {
			delta = delta + fmt.Sprintf("K %s %s %s\n", string(change.byte), tag, change.string)
		}
	}
	if len(L) > 0 {
		for _, l := range L {
			delta = delta + fmt.Sprintf("L %s\n", l)
		}
	}
	delta = delta + fmt.Sprintf("P %s\nU %s\n", P, get_user())
	Z, err := hash([]byte(delta))
	if err != nil {
		return nil, err
	}
	delta = delta + fmt.Sprintf(
		"Z %s", Z)

	return []byte(delta), nil
}

func get_root_ref(db *sql.DB, I string) (string, error) {
	var id string
	query := "SELECT id FROM ref WHERE uid = ?;"
	row, err := db.Query(query, I)
	if err != nil {
		return "", err
	}
	row.Next()
	row.Scan(&id)
	return id, nil

}

func get_previous_delta(db *sql.DB, I string) (string, error) {
	var P string
	query := `
		SELECT d.id FROM ref_delta d 
		JOIN ref r ON r.id = d.ref_id
		WHERE r.uid = ? ORDER BY d.ctime DESC
		LIMIT 1;
		`
	row, err := db.Query(query, I)

	if err != nil {
		return "", err
	}
	row.Next()
	row.Scan(&P)
	return P, nil
}

func get_ref_history() {}
