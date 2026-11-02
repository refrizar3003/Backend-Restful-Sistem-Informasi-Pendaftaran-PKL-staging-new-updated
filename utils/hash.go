package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword menghasilkan bcrypt hash dari password plaintext.
// Menggunakan DefaultCost (10) sebagai cost factor.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash memvalidasi password plaintext terhadap bcrypt hash.
// Return true jika password cocok, false jika tidak.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}