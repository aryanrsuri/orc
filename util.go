package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

func read_to_bytes(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}

	return data, nil
}

func get_user() string {
	U, err := user.Current()
	if err != nil {
		return "default"
	}
	return U.Username
}

func uuid() string {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[:8], uint64(time.Now().Unix()))
	rand.Read(b[8:])
	return hex.EncodeToString(b)[6:]
}

func walk_dir(root string, ignore []string) ([]string, error) {
	var files []string
	ignored := make(map[string]bool)
	for _, s := range ignore {
		ignored[s] = true
	}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && ignored[d.Name()] {
			return filepath.SkipDir
		}

		if d.IsDir() || ignored[d.Name()] {
			return nil
		}
		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
