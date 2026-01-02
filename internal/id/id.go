package id

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/oklog/ulid/v2"
)

// GenerateULID generates a new ULID string.
// It uses crypto/rand for entropy.
func GenerateULID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

// GetHashID returns the first 8 characters of the SHA-256 hash of the input string.
func GetHashID(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])[:8]
}
