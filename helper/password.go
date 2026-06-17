package helper

import (
	"golang.org/x/crypto/bcrypt"
	"os"
	"strconv"
)

// HashPassword hashes a password with bcrypt.
func HashPassword(plain string) (string, error) {
	cost := bcrypt.DefaultCost
	if v := os.Getenv("BCRYPT_COST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= bcrypt.MinCost && n <= bcrypt.MaxCost {
			cost = n
		}
	}
	b, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword compares bcrypt hash with plain password.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
