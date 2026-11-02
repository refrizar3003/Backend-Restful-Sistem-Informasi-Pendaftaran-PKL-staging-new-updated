package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateDeviceToken menghasilkan token acak 32 byte (64 karakter hex)
func GenerateDeviceToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashDeviceToken mengembalikan SHA-256 hex dari token, untuk disimpan di database
func HashDeviceToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}