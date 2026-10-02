package service

import (
	"strings"
	"time"

	"atchannel-backend/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
)

// newClaims builds claims with controlled timestamps for tests that need a
// token that is already expired.
func newClaims(userID, tokenType string, expiry time.Time) middleware.UserClaims {
	return middleware.UserClaims{
		UserID:    userID,
		Email:     "a@example.com",
		Roles:     []string{"user"},
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "atchannel",
			Subject:   userID,
		},
	}
}

// splitToken cuts a compact JWT into its three base64url segments.
func splitToken(token string) [3]string {
	var out [3]string
	parts := strings.Split(token, ".")
	copy(out[:], parts)
	return out
}
