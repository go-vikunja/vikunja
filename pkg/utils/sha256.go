package utils

import (
	"crypto/sha256"
	"fmt"
)

// Sha256 calculates a sha256 hash from a string
func Sha256(cleartext string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(cleartext)))[:45]
}
