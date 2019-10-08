package test

import (
	"fmt"
)

func DiffBytes(str1 string, str2 string) {
	b1 := []byte(str1)
	b2 := []byte(str2)
	if len(b1) != len(b2) {
		fmt.Printf("Different lengths: %d %d\n", len(b1), len(b2))
	}
	for i, b := range b1 {
		if i < len(b2) {
			if b != b2[i] {
				fmt.Printf("byte %d differs: %d, %d\n", i, b, b2[i])
			}
		}
	}
}

// compare human readable error message, ignoring wrapped objects
func EqualMessage(e1 error, e2 error) bool {
	if e1 == nil && e2 == nil {
		return true
	}
	if e1 != nil && e2 != nil {
		return e1.Error() == e2.Error()
	}
	return false
}
