// Parse control artifacts into their respective set of fields and ensure validitity
package main

import (
	"fmt"
	"strings"
)

type c_artifact struct {
	c_type string
	fields []map[string][]string
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

func parse_artifact(b *blob) (*c_artifact, error) {
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

	return &control, nil
}
