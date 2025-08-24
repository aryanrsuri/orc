package main

import (
	"database/sql"
	"fmt"
	"os/user"
	"time"
)

func put_manifest(db *sql.DB, root string, ignore []string) (string, error) {
	manifest := "A M\n"
	manifest = manifest + fmt.Sprintf("D %d\n", time.Now().Unix())
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	manifest = manifest + fmt.Sprintf("U %s\n", user.Username)
	files, err := walk_dir(root, ignore)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		data, err := read_to_bytes(f)
		if err !=nil {
			return "", err
		}
		id, err := put_blob(db, data)
		if err !=nil {
			return "", err
		}
		manifest = manifest + fmt.Sprintf("F %s %s\n", f, id)
	}
	// Get previous manifest id

	// TODO: Create manifest entry, and insert files into manifest_file given the manifest id
	fmt.Println(manifest)
	return "", nil
}
