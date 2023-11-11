package testutils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomString generates a random string
func GenerateRandomString() string {
	buf := make([]byte, 16)

	n, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(buf[:n])
}
