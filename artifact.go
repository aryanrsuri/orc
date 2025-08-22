// Parse control artifacts into their respective set of fields and ensure validitity
package main

import (
	"fmt"
	"strings"
)


func parse_artifact (b *blob) {
	content := string(b.data)
	fmt.Println(content)
	fmt.Println(strings.Split(content, "\n"))
}
