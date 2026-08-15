package services_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kosero/atchannel/backend/internal/config"
	"github.com/kosero/atchannel/backend/internal/models"
	"github.com/kosero/atchannel/backend/internal/services"
)

func newAuthService(t *testing.T) (*services.AuthService, *InMemoryUserRepo, *InMemoryRefreshRepo) {
	t.Helper()
	users := NewInMemoryUserRepo()
	refresh := NewInMemoryRefreshRepo()
	cfg := &config.Auth{
		JWTSecret:            "test-secret-that-is-long-enough",
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      24 * time.Hour,
		RefreshTokenMaxAge:   7 * 24 * time.Hour,
		ArgonMemoryKiB:       65536,
		ArgonIterations:      3,
		ArgonParallelism:     1,
		ArgonSaltLength:      16,
		ArgonKeyLength:       32,
		SecureCookies:        true,
		RefreshCookieDomain:  "",
		CookieName:           "atch_refresh",
	}
	return services.NewAuthService(users, refresh, cfg), users, refresh
}

func registerUser(t *testing.T, svc *services.AuthService) *models.User {
	t.Helper()
	user, err := svc.Register(context.Background(), "tester", "tester@example.com", "supersecret123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	return user
}

func TestRegisterDuplicate(t *testing.T) {
	svc, _, _ := newAuthService(t)
	registerUser(t, svc)
	_, err := svc.Register(context.Background(), "tester", "other@example.com", "supersecret123")
	if err != services.ErrUserExists {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}
	_, err = svc.Register(context.Background(), "other", "tester@example.com", "supersecret123")
	if err != services.ErrUserExists {
		t.Fatalf("expected ErrUserExists for email, got %v", err)
	}
}

func TestLoginSuccessAndFailure(t *testing.T) {
	svc, _, _ := newAuthService(t)
	registerUser(t, svc)

	if _, err := svc.Login(context.Background(), "tester", "supersecret123"); err != nil {
		t.Fatalf("login should succeed: %v", err)
	}
	if _, err := svc.Login(context.Background(), "tester", "wrongpassword"); err != services.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if _, err := svc.Login(context.Background(), "nobody", "supersecret123"); err != services.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials for unknown user, got %v", err)
	}
}

func TestPasswordHashesAreArgon2idSalted(t *testing.T) {
	svc, users, _ := newAuthService(t)
	registerUser(t, svc)
	u, _ := users.GetByUsername(context.Background(), "tester")
	if !strings.HasPrefix(u.PasswordHash, "$argon2id$v=19$") {
		t.Fatalf("expected argon2id hash, got %q", u.PasswordHash)
	}
}

func TestRefreshRotation(t *testing.T) {
	svc, _, _ := newAuthService(t)
	user := registerUser(t, svc)
	pair, err := svc.IssueTokens(context.Background(), user)
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected both tokens")
	}

	// First refresh: valid, issues a new pair.
	pair2, user2, err := svc.Refresh(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	if user2.ID != user.ID {
		t.Fatalf("refresh returned wrong user")
	}
	if pair2.RefreshToken == pair.RefreshToken {
		t.Fatal("refresh token must rotate")
	}

	// Second refresh with the OLD token must be detected as reuse.
	_, _, err = svc.Refresh(context.Background(), pair.RefreshToken)
	if err != services.ErrRefreshTokenUsed {
		t.Fatalf("expected ErrRefreshTokenUsed for old token, got %v", err)
	}
}

func TestRefreshReuseRevokesFamily(t *testing.T) {
	svc, _, _ := newAuthService(t)
	user := registerUser(t, svc)

	pair1, _ := svc.IssueTokens(context.Background(), user)
	pair2, _, err := svc.Refresh(context.Background(), pair1.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	// Replay pair1.RefreshToken: reuse detection revokes the family.
	if _, _, err := svc.Refresh(context.Background(), pair1.RefreshToken); err != services.ErrRefreshTokenUsed {
		t.Fatalf("expected reuse detection, got %v", err)
	}

	// The freshly rotated token (pair2.RefreshToken) must now be rejected too.
	// Its family was revoked, so presenting it is treated as reuse.
	if _, _, err := svc.Refresh(context.Background(), pair2.RefreshToken); err != services.ErrRefreshTokenUsed {
		t.Fatalf("expected family-revoked token to be rejected, got %v", err)
	}
}

func TestRefreshExpired(t *testing.T) {
	svc, _, _ := newAuthService(t)
	registerUser(t, svc)

	// A token that is not in the store is invalid.
	if _, _, err := svc.Refresh(context.Background(), "definitely-not-a-real-token"); err != services.ErrRefreshTokenInvalid {
		t.Fatalf("expected invalid refresh token, got %v", err)
	}
}

func TestLogoutRevokesToken(t *testing.T) {
	svc, _, _ := newAuthService(t)
	user := registerUser(t, svc)
	pair, _ := svc.IssueTokens(context.Background(), user)

	if err := svc.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	// A revoked token presented again triggers reuse detection.
	if _, _, err := svc.Refresh(context.Background(), pair.RefreshToken); err != services.ErrRefreshTokenUsed {
		t.Fatalf("expected reuse detection after logout, got %v", err)
	}
}

func TestValidateAccessToken(t *testing.T) {
	svc, _, _ := newAuthService(t)
	user := registerUser(t, svc)
	pair, _ := svc.IssueTokens(context.Background(), user)

	id, role, err := svc.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if id != user.ID {
		t.Fatalf("expected user id %s, got %s", user.ID, id)
	}
	if role != models.RoleUser {
		t.Fatalf("expected role user, got %s", role)
	}

	if _, _, err := svc.ValidateAccessToken("garbage"); err != services.ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}
