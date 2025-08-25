package main

import (
	"database/sql"
	"fmt"
	"time"
)

// FIXME: SQL should operate on one commit thread here, rolling back on any error
func put_manifest(db *sql.DB, root string, ignore []string) (string, error) {
	F := make(map[string]string)
	files, err := walk_dir(root, ignore)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		data, err := read_to_bytes(f)
		if err != nil {
			return "", err
		}
		id, err := put_blob(db, data)
		if err != nil {
			return "", err
		}
		F[f] = id
	}
	// Get previous manifest id
	P, err := get_previous_manifest(db)
	if err != nil {
		return "", err
	}

	// TODO: Create manifest entry, and insert files into manifest_file given the manifest id
	// FIXME: Should everything below, not just be in ``write_manifest``
	manifest, err := write_manifest(F, P)
	if err != nil {
		return "", err
	}
	id, err := put_blob(db, manifest)
	if err != nil {
		return "", err
	}

	insert_manifest := "INSERT INTO manifest (id, ctime, user) VALUES (?, ?, ?)"
	insert_manifest_file := "INSERT INTO manifest_file (m_id, path, blob) VALUES (?, ?, ?)"

	_, err = db.Exec(insert_manifest, id, time.Now().Unix(), get_user())
	if err != nil {
		return "", nil
	}

	for f, b := range F {
		_, err = db.Exec(insert_manifest_file, id, f, b)
		if err != nil {
			return "", nil
		}
	}
	return id, nil
}

func get_previous_manifest(db *sql.DB) (string, error) {
	query := "SELECT MAX(id) as 'id' from manifest;"
	var id string
	row, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	row.Next()
	row.Scan(&id)
	return id, nil
}

func write_manifest(F map[string]string, P string) ([]byte, error) {

	manifest := fmt.Sprintf(
		"A M\nD %d\n", time.Now().Unix())
	for f, b := range F {
		manifest = manifest + fmt.Sprintf(
			"F %s %s\n", f, b)
	}
	if len(P) > 0 {
		manifest = manifest + fmt.Sprintf(
			"P %s\nU %s\n", P, get_user())
	}
	Z, err := hash([]byte(manifest))
	if err != nil {
		return nil, err
	}
	manifest = manifest + fmt.Sprintf(
		"Z %s", Z)
	return []byte(manifest), nil
}
