package helper

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwtIssuer = "oryoo.com/broomax"

// AdminJWTClaims carries identity for admin JWTs.
type AdminJWTClaims struct {
	AdminID int64  `json:"admin_id"`
	Email   string `json:"email"`
	jwt.RegisteredClaims
}

// SignAdminJWT issues a short-lived HS256 JWT for an authenticated admin.
func SignAdminJWT(adminID int64, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET is not set")
	}
	ttl := 24 * time.Hour
	if v := os.Getenv("JWT_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttl = time.Duration(n) * time.Second
		}
	}
	now := time.Now()
	claims := AdminJWTClaims{
		AdminID: adminID,
		Email:   email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			Subject:   fmt.Sprintf("%d", adminID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

// ParseAndVerifyAdminJWT validates signature, expiry, and issuer.
func ParseAndVerifyAdminJWT(tokenStr string) (*AdminJWTClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is not set")
	}
	tok, err := jwt.ParseWithClaims(tokenStr, &AdminJWTClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*AdminJWTClaims)
	if !ok || !tok.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	if claims.Issuer != jwtIssuer {
		return nil, fmt.Errorf("invalid issuer")
	}
	return claims, nil
}
