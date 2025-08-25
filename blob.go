package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
)

const DATA_MINUS_C_CARD int = 66

type blob struct {
	id    string
	ctime int64
	size  int64
	data  []byte
}

type pair struct {
	byte
	string
}

type c_artifact struct {
	A byte
	C string
	D int64
	F map[string]string
	I string
	K map[string]pair
	L []string
	P string
	T string
	U string
	Z string
}

func is_valid_hash(db *sql.DB, data string) bool {
	if len(data) != 64 {
		return false
	}

	// TODO: This is If I want to check the validity of a valid_hash
	// in general... Should remove?
	if db != nil {
		query := "SELECT 1 FROM blob WHERE id = ?;"
		rows, err := db.Query(query, data)
		if err != nil {
			panic(err)
		}
		if !rows.Next() {
			return false
		}
	}

	return true
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

	// FIXME: Should I first check, e.g. ``SELECT 1 ... id = id``, or keep this?...
	if err != nil {
		if err.(sqlite3.Error).ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
			return id, nil
		}
		return "", err
	}
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

// TODO: Check “field“ length on per-card basis
// e.g. if “F“ contains field length != 2, panic
func get_card(line string) (byte, []string, error) {
	parts := strings.Fields(line)

	if len(parts) == 0 {
		return 0, nil, fmt.Errorf("empty line")
	}
	if len(parts) == 1 {
		return 0, nil, fmt.Errorf("empty card")
	}
	head := parts[0][0]
	field := parts[1:]
	return head, field, nil
}

func get_artifact(db *sql.DB, b *blob) (*c_artifact, error) {
	var control c_artifact

	control.F = make(map[string]string)
	control.K = make(map[string]pair)
	content := string(b.data)
	checksum, err := hash(b.data[:len(b.data)-DATA_MINUS_C_CARD])
	if err != nil {
		return nil, err
	}
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		c, field, err := get_card(line)
		if err != nil {
			return nil, err
		}
		// FIXME: Push error handling to ``get_card``
		switch c {
		default:
			return nil, fmt.Errorf("impossible card type %s", string(c))
		case 'A':
			control.A = field[0][0]
		case 'C':
			control.C = strings.Join(field, " ")
		case 'D':
			{
				ts, err := strconv.ParseInt(field[0], 10, 0)
				if err != nil {
					return nil, err
				}
				control.D = ts
			}
		case 'F':
			{
				// FIXME: Enforce ``is_valid_hash``
				/*
					if !is_valid_hash(db, field[1]) {
						return nil, fmt.Errorf("F card does not contain a valid artifact id")
					}
				*/
				control.F[field[0]] = field[1]
			}
		case 'I':
			control.I = field[0]
		case 'K':
			control.K[field[1]] = pair{field[0][0], field[2]}
		case 'L':
			{
				// FIXME: Enforce ``is_valid_hash``
				/*
					if !is_valid_hash(db, field[1]) {
						return nil, fmt.Errorf("L card does not contain a valid link id")
					}
				*/
				control.L = append(control.L, field[0])
			}
		case 'P':
			{
				// FIXME: Enforce ``is_valid_hash``
				/*
					if !is_valid_hash(db, field[1]) {
						return nil, fmt.Errorf("P card does not contain a valid link id")
					}
				*/
				control.P = field[0]
			}
		case 'T':
			control.T = field[0]
		case 'U':
			control.U = field[0]
		case 'Z':
			{
				if checksum != field[0] {
					return nil, fmt.Errorf("Checksum does not match")
				}
				control.Z = field[0]
			}
		}
	}

	return &control, nil
}

func valid_artifact(control *c_artifact) bool {
	result := true
	switch control.A {
	case 'D', 'M', 'C':
		panic("Not implemented")
	case 'R':
		if len(control.F) > 0 || control.P != "" {
			result = false
		}

	}
	return result
}
