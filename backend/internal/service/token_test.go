package service

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenPairRoundTrip(t *testing.T) {
	svc := NewTokenService([]byte("test-secret"), "atchannel")

	pair, err := svc.GenerateTokenPair("42", "a@example.com", []string{"user", "admin"})
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	access, err := svc.ValidateToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken(access): %v", err)
	}
	if access.UserID != "42" || access.Email != "a@example.com" || access.TokenType != "access" {
		t.Errorf("access claims = %+v, want user 42 / a@example.com / access", access)
	}
	if len(access.Roles) != 2 || access.Roles[1] != "admin" {
		t.Errorf("access roles = %v, want [user admin]", access.Roles)
	}

	refresh, err := svc.ValidateRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("ValidateRefreshToken(refresh): %v", err)
	}
	if refresh.TokenType != "refresh" || refresh.UserID != "42" {
		t.Errorf("refresh claims = %+v, want refresh / 42", refresh)
	}
}

// An access token must never be usable as a refresh token (and vice versa):
// that is what stops a leaked short-lived token from minting new ones.
func TestTokenTypeIsEnforced(t *testing.T) {
	svc := NewTokenService([]byte("test-secret"), "atchannel")

	pair, err := svc.GenerateTokenPair("1", "a@example.com", []string{"user"})
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	if _, err := svc.ValidateRefreshToken(pair.AccessToken); !errors.Is(err, ErrWrongType) {
		t.Errorf("ValidateRefreshToken(access) = %v, want ErrWrongType", err)
	}
	if _, err := svc.ValidateToken(pair.RefreshToken); err != nil {
		t.Errorf("ValidateToken(refresh) = %v, want nil (type is only checked by the refresh path)", err)
	}
}

func TestTokenSignedWithOtherSecretIsRejected(t *testing.T) {
	good := NewTokenService([]byte("test-secret"), "atchannel")
	other := NewTokenService([]byte("different-secret"), "atchannel")

	pair, err := other.GenerateTokenPair("1", "a@example.com", []string{"user"})
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	if _, err := good.ValidateToken(pair.AccessToken); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("ValidateToken with wrong secret = %v, want ErrInvalidToken", err)
	}
}

func TestExpiredTokenIsRejected(t *testing.T) {
	svc := NewTokenService([]byte("test-secret"), "atchannel")

	claims := newClaims("1", "refresh", time.Now().Add(-time.Minute))
	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := expired.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := svc.ValidateRefreshToken(signed); !errors.Is(err, ErrExpiredToken) {
		t.Errorf("ValidateRefreshToken(expired) = %v, want ErrExpiredToken", err)
	}
}

// Tampering with any part of the token — payload included — must break the
// signature check.
func TestTamperedTokenIsRejected(t *testing.T) {
	svc := NewTokenService([]byte("test-secret"), "atchannel")

	pair, err := svc.GenerateTokenPair("1", "a@example.com", []string{"user"})
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	// Flip a character in the payload segment (the middle one).
	parts := splitToken(pair.AccessToken)
	payload := []byte(parts[1])
	if payload[0] == 'A' {
		payload[0] = 'B'
	} else {
		payload[0] = 'A'
	}
	tampered := parts[0] + "." + string(payload) + "." + parts[2]

	if _, err := svc.ValidateToken(tampered); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("ValidateToken(tampered) = %v, want ErrInvalidToken", err)
	}
}
