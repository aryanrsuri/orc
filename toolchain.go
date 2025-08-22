// CLI toolchain
package main

import (
	"os"
)

func read_to_bytes(path string) ([]byte, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}

	return data, nil
}
