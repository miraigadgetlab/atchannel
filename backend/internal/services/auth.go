package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"

	"github.com/kosero/atchannel/backend/internal/config"
	"github.com/kosero/atchannel/backend/internal/models"
	"github.com/kosero/atchannel/backend/internal/repositories"
	"github.com/kosero/atchannel/backend/internal/repositories/postgres"
)

var defaultColors = []string{
	"#7C3AED",
	"#3B82F6",
	"#10B981",
	"#8B5CF6",
	"#06B6D4",
	"#059669",
	"#6366F1",
	"#0EA5E9",
	"#14B8A6",
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrRefreshTokenUsed   = errors.New("refresh token reuse detected")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrRefreshTokenInvalid = errors.New("invalid refresh token")
)

type AuthService struct {
	users  repositories.UserRepository
	tokens repositories.RefreshTokenRepository
	cfg    *config.Auth
	now    func() time.Time

	// dummy hash used to equalize Argon2 timing when a login targets a
	// non-existent username, preventing user enumeration by response time.
	dummyOnce sync.Once
	dummyHash string
	dummyErr  error
}

func NewAuthService(users repositories.UserRepository, tokens repositories.RefreshTokenRepository, cfg *config.Auth) *AuthService {
	return &AuthService{users: users, tokens: tokens, cfg: cfg, now: time.Now}
}

// ensureDummyHash computes a throwaway Argon2 hash (with the current
// parameters) once; verifying a password against it costs the same CPU as a
// real verification.
func (s *AuthService) ensureDummyHash() {
	s.dummyOnce.Do(func() {
		s.dummyHash, s.dummyErr = s.hashPassword("dummy-password-for-timing")
	})
}

// --- Password hashing (Argon2id) ---

type argonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  int
	keyLength   uint32
}

func (s *AuthService) params() argonParams {
	return argonParams{
		memory:      s.cfg.ArgonMemoryKiB,
		iterations:  s.cfg.ArgonIterations,
		parallelism: s.cfg.ArgonParallelism,
		saltLength:  s.cfg.ArgonSaltLength,
		keyLength:   s.cfg.ArgonKeyLength,
	}
}

func (s *AuthService) hashPassword(password string) (string, error) {
	p := s.params()
	salt := make([]byte, p.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("salt generation: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)
	encoded := base64.RawStdEncoding.EncodeToString(key)
	saltEnc := base64.RawStdEncoding.EncodeToString(salt)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", p.memory, p.iterations, p.parallelism, saltEnc, encoded), nil
}

func (s *AuthService) verifyPassword(password, encodedHash string) (bool, error) {
	// Format: $argon2id$v=19$m=<mem>,t=<time>,p=<parallel>$<salt>$<key>
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, fmt.Errorf("parse hash: unsupported format")
	}
	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("parse hash params: %w", err)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode key: %w", err)
	}
	derived := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expected)))
	return subtle.ConstantTimeCompare(derived, expected) == 1, nil
}

// --- Register / Login ---

func (s *AuthService) Register(ctx context.Context, username, email, password string) (*models.User, error) {
	if username == "" || email == "" || password == "" {
		return nil, fmt.Errorf("username, email and password are required")
	}
	exists, err := s.users.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}
	exists, err = s.users.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}
	hash, err := s.hashPassword(password)
	if err != nil {
		return nil, err
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(defaultColors))))
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Role:         models.RoleUser,
		Color:        defaultColors[n.Int64()],
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*models.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			// Burn an Argon2 verification so the response time matches a
			// real (failed) login; otherwise the timing difference reveals
			// whether the username exists.
			s.ensureDummyHash()
			if s.dummyErr == nil {
				_, _ = s.verifyPassword(password, s.dummyHash)
			}
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	ok, err := s.verifyPassword(password, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

// --- Tokens ---

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

func (s *AuthService) IssueTokens(ctx context.Context, user *models.User) (*TokenPair, error) {
	access, err := s.issueAccessToken(user)
	if err != nil {
		return nil, err
	}
	refresh, err := s.issueRefreshToken(ctx, user)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) issueAccessToken(user *models.User) (string, error) {
	now := s.now()
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": string(user.Role),
		"jti":  newID(),
		"iat":  now.Unix(),
		"exp":  now.Add(s.cfg.AccessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) issueRefreshToken(ctx context.Context, user *models.User) (string, error) {
	family := newID()
	return s.issueRefreshTokenInFamily(ctx, user, family)
}

func (s *AuthService) issueRefreshTokenInFamily(ctx context.Context, user *models.User, family string) (string, error) {
	raw := newID() + ":" + newID()
	hash := hashToken(raw)
	expiresAt := s.now().Add(s.cfg.RefreshTokenTTL)
	if _, err := s.tokens.Create(ctx, user.ID, family, hash, expiresAt); err != nil {
		return "", err
	}
	return raw, nil
}

func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*TokenPair, *models.User, error) {
	hash := hashToken(rawToken)
	stored, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, nil, ErrRefreshTokenInvalid
		}
		return nil, nil, err
	}

	now := s.now()
	if stored.Revoked {
		// Reuse detection: a rotated token presented again means theft.
		if err := s.tokens.RevokeFamily(ctx, stored.FamilyID); err != nil {
			return nil, nil, err
		}
		return nil, nil, ErrRefreshTokenUsed
	}
	if stored.ExpiresAt.Before(now) {
		if err := s.tokens.Revoke(ctx, stored.ID, nil); err != nil {
			return nil, nil, err
		}
		return nil, nil, ErrRefreshTokenExpired
	}

	user, err := s.users.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, nil, err
	}

	// Rotate: revoke current, issue new in same family.
	newRaw, err := s.issueRefreshTokenInFamily(ctx, user, stored.FamilyID)
	if err != nil {
		return nil, nil, err
	}
	if err := s.tokens.Revoke(ctx, stored.ID, nil); err != nil {
		return nil, nil, err
	}

	access, err := s.issueAccessToken(user)
	if err != nil {
		return nil, nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: newRaw,
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, user, nil
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	hash := hashToken(rawToken)
	stored, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil // already gone
		}
		return err
	}
	return s.tokens.Revoke(ctx, stored.ID, nil)
}

// ValidateAccessToken parses and validates an HS256 access token, returning
// the user ID and role. Returns ErrInvalidCredentials for any failure.
func (s *AuthService) ValidateAccessToken(tokenString string) (userID string, role models.Role, err error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", "", ErrInvalidCredentials
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", ErrInvalidCredentials
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", "", ErrInvalidCredentials
	}
	roleStr, _ := claims["role"].(string)
	return sub, models.Role(roleStr), nil
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("entropy failure: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
