package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"atchannel-backend/internal/models"

	"gorm.io/gorm"
)

// SessionService keeps refresh tokens revocable. Tokens are stored as SHA-256
// hashes, checked against expiry on every refresh, and rotated when used.
type SessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

// HashToken returns the storage form of a refresh token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Record stores a freshly issued refresh token and prunes this user's
// already expired rows so the table does not grow forever.
func (s *SessionService) Record(ctx context.Context, userID uint, refreshToken string, expiresAt time.Time) error {
	s.db.WithContext(ctx).
		Where("user_id = ? AND expires_at < ?", userID, time.Now()).
		Delete(&models.RefreshToken{})

	row := models.RefreshToken{
		UserID:    userID,
		TokenHash: HashToken(refreshToken),
		ExpiresAt: expiresAt,
	}

	return s.db.WithContext(ctx).Create(&row).Error
}

// IsValid reports whether the token exists, is neither revoked nor expired.
func (s *SessionService) IsValid(ctx context.Context, refreshToken string) (bool, error) {
	var count int64

	err := s.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", HashToken(refreshToken), time.Now()).
		Count(&count).Error

	return count > 0, err
}

// Revoke invalidates a single session. Returns false when the token was
// unknown or already revoked, which keeps logout idempotent.
func (s *SessionService) Revoke(ctx context.Context, refreshToken string) (bool, error) {
	res := s.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", HashToken(refreshToken)).
		Update("revoked_at", time.Now())

	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// RevokeAllForUser kills every session of a user (logout everywhere,
// password change, forced sign-out). Returns how many sessions died.
func (s *SessionService) RevokeAllForUser(ctx context.Context, userID uint) (int64, error) {
	res := s.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", time.Now())

	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
