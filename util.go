package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
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

type state struct {
	U []string
	M []string
	D []string
}

func format_state(s *state) string {
	var result string
	for _, u := range s.U {
		result = result + fmt.Sprintf("U %s\n", u)
	}
	for _, u := range s.M {
		result = result + fmt.Sprintf("M %s\n", u)
	}
	for _, u := range s.D {
		result = result + fmt.Sprintf("D %s\n", u)
	}
	return result
}

func state_clean(s *state) bool {
	if s.U == nil && s.D == nil && s.M == nil {
		return true
	}
	return false
}
