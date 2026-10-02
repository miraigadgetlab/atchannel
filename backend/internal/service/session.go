package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"atchannel-backend/internal/models"

	"gorm.io/gorm"
)

// ErrTokenReuse means a refresh token was presented after it had already
// been revoked. Either the token was stolen, or a client replayed an old
// one; there is no way to tell which, so the safe answer is to assume the
// worst and kill the whole session family.
var ErrTokenReuse = errors.New("refresh token reuse detected")

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

// NewFamilyID mints a fresh identifier for one login's token lineage.
func NewFamilyID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// crypto/rand failing is not recoverable; a timestamp-derived id is
		// still unique enough for grouping and never blocks a login.
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(buf[:])
}

// Record stores a freshly issued refresh token and prunes this user's
// already expired rows so the table does not grow forever.
func (s *SessionService) Record(ctx context.Context, userID uint, familyID, refreshToken string, expiresAt time.Time) error {
	s.db.WithContext(ctx).
		Where("user_id = ? AND expires_at < ?", userID, time.Now()).
		Delete(&models.RefreshToken{})

	row := models.RefreshToken{
		UserID:    userID,
		FamilyID:  familyID,
		TokenHash: HashToken(refreshToken),
		ExpiresAt: expiresAt,
	}

	return s.db.WithContext(ctx).Create(&row).Error
}

// Lookup returns the stored row for a presented refresh token, or nil when
// unknown.
func (s *SessionService) Lookup(ctx context.Context, refreshToken string) (*models.RefreshToken, error) {
	var row models.RefreshToken

	err := s.db.WithContext(ctx).
		Where("token_hash = ?", HashToken(refreshToken)).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &row, nil
}

// Rotate revokes the presented token and admits its replacement, but only
// if the presented one is still live. Everything happens in one transaction
// so a crash cannot leave the user with a token that is neither revoked nor
// recorded.
//
// A token that is already revoked triggers reuse detection: the attacker is
// replaying something we issued earlier, so the entire family dies and both
// parties are signed out.
//
// The burn-outcome is signalled through a flag rather than by returning an
// error from the transaction callback, because GORM rolls back whatever the
// callback wrote when it returns an error — returning ErrTokenReuse from in
// there would silently undo the very revocation that makes reuse detection
// work. The callback therefore commits, and the error is raised afterwards.
func (s *SessionService) Rotate(ctx context.Context, oldRefreshToken, newRefreshToken string, newExpiresAt time.Time) (*models.RefreshToken, error) {
	var result *models.RefreshToken
	var outcome error

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row models.RefreshToken

		err := tx.Where("token_hash = ?", HashToken(oldRefreshToken)).
			First(&row).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Unknown token: nothing to burn, but the caller must not
				// treat this as a valid session either. There is nothing to
				// commit here, so committing is a no-op.
				outcome = ErrTokenReuse
				return nil
			}
			return err
		}

		if row.RevokedAt != nil {
			// Reuse of a dead token — burn the whole family, then report it.
			if err := tx.Model(&models.RefreshToken{}).
				Where("family_id = ?", row.FamilyID).
				Updates(map[string]any{
					"revoked_at":  time.Now(),
					"revoked_for": "reuse",
				}).Error; err != nil {
				return err
			}
			outcome = ErrTokenReuse
			return nil
		}

		if !row.ExpiresAt.After(time.Now()) {
			outcome = ErrExpiredToken
			return nil
		}

		now := time.Now()
		if err := tx.Model(&models.RefreshToken{}).
			Where("id = ?", row.ID).
			Updates(map[string]any{
				"revoked_at":  now,
				"revoked_for": "rotated",
			}).Error; err != nil {
			return err
		}

		replacement := models.RefreshToken{
			UserID:    row.UserID,
			FamilyID:  row.FamilyID,
			TokenHash: HashToken(newRefreshToken),
			ExpiresAt: newExpiresAt,
		}
		if err := tx.Create(&replacement).Error; err != nil {
			return err
		}

		result = &replacement
		return nil
	})
	if err != nil {
		return nil, err
	}
	if outcome != nil {
		return nil, outcome
	}

	return result, nil
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
func (s *SessionService) Revoke(ctx context.Context, refreshToken, reason string) (bool, error) {
	res := s.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", HashToken(refreshToken)).
		Updates(map[string]any{
			"revoked_at":  time.Now(),
			"revoked_for": reason,
		})

	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// RevokeAllForUser kills every session of a user (logout everywhere,
// password change, forced sign-out). Returns how many sessions died.
func (s *SessionService) RevokeAllForUser(ctx context.Context, userID uint, reason string) (int64, error) {
	res := s.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]any{
			"revoked_at":  time.Now(),
			"revoked_for": reason,
		})

	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
