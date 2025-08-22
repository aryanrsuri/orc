// Parse control artifacts into their respective set of fields and ensure validitity
package main

import "fmt"


func parse_object (b *blob) {
	content := string(b.data)
	fmt.Println(content)
}
