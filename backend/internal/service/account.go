package service

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"strings"
	"time"

	"atchannel-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RecoveryCodeCount is how many single-use codes one batch contains.
const RecoveryCodeCount = 10

// recoveryAlphabet omits the characters that are easy to confuse when a code
// is written down or read back (0/O, 1/I/L).
const recoveryAlphabet = "23456789ABCDEFGHJKMNPQRSTVWXYZ"

var ErrInvalidRecoveryCode = errors.New("recovery code is invalid or has already been used")

// AccountService owns the account's credential recovery: issuing and
// consuming single-use recovery codes. Only hashes are ever stored.
type AccountService struct {
	db       *gorm.DB
	users    *UserService
	sessions *SessionService
}

func NewAccountService(db *gorm.DB, users *UserService, sessions *SessionService) *AccountService {
	return &AccountService{
		db:       db,
		users:    users,
		sessions: sessions,
	}
}

// ResetPasswordWithCode is the account recovery path: a single-use recovery
// code stands in for a reset link. The name is only used to look the account
// up, so a wrong code can never reset somebody else's password.
func (s *AccountService) ResetPasswordWithCode(ctx context.Context, name, code, newPassword string) error {
	user, err := s.users.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return ErrInvalidRecoveryCode
		}
		return err
	}

	if err := s.consumeRecoveryCode(ctx, user.ID, code); err != nil {
		return err
	}

	return s.setPassword(ctx, user.ID, newPassword)
}

// GenerateRecoveryCodes replaces every outstanding code for the account with a
// fresh batch and returns the plaintext codes exactly once. Only hashes are
// stored, so a lost copy cannot be re-read from the database.
func (s *AccountService) GenerateRecoveryCodes(ctx context.Context, userID uint) ([]string, error) {
	codes := make([]string, 0, RecoveryCodeCount)
	rows := make([]models.RecoveryCode, 0, RecoveryCodeCount)

	for i := 0; i < RecoveryCodeCount; i++ {
		code, err := randomRecoveryCode()
		if err != nil {
			return nil, err
		}
		codes = append(codes, code)
		rows = append(rows, models.RecoveryCode{
			UserID:    userID,
			TokenHash: HashToken(normalizeRecoveryCode(code)),
		})
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.RecoveryCode{}).Error; err != nil {
			return err
		}
		return tx.Create(&rows).Error
	})
	if err != nil {
		return nil, err
	}

	return codes, nil
}

// RecoveryCodeStatus reports how many unused recovery codes the account has
// left, so the settings page can nudge the owner to regenerate them.
func (s *AccountService) RecoveryCodeStatus(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.RecoveryCode{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// setPassword stores a new password hash and revokes every session.
func (s *AccountService) setPassword(ctx context.Context, userID uint, newPassword string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("password", string(hashed)).Error; err != nil {
		return err
	}

	if _, err := s.sessions.RevokeAllForUser(ctx, userID, "password-reset"); err != nil {
		log.Printf("could not revoke sessions after reset for user %d: %v", userID, err)
	}
	return nil
}

// consumeRecoveryCode validates and burns one single-use code for the account.
// Recovery codes never expire: they stay valid until they are used or the
// owner regenerates the batch.
func (s *AccountService) consumeRecoveryCode(ctx context.Context, userID uint, code string) error {
	var outcome error

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row models.RecoveryCode

		// FOR UPDATE so two concurrent redemptions of the same code cannot
		// both succeed.
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND token_hash = ? AND used_at IS NULL",
				userID, HashToken(normalizeRecoveryCode(code))).
			First(&row).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				outcome = ErrInvalidRecoveryCode
				return nil
			}
			return err
		}

		return tx.Model(&row).Update("used_at", time.Now()).Error
	})
	if err != nil {
		return err
	}

	return outcome
}

// randomRecoveryCode mints one human-readable code in XXXXX-XXXXX form.
func randomRecoveryCode() (string, error) {
	const groups = 2
	const groupLen = 5

	var b strings.Builder
	b.Grow(groups*groupLen + groups - 1)

	for i := 0; i < groups*groupLen; i++ {
		if i > 0 && i%groupLen == 0 {
			b.WriteByte('-')
		}
		c, err := randomAlphabetChar(recoveryAlphabet)
		if err != nil {
			return "", err
		}
		b.WriteByte(c)
	}

	return b.String(), nil
}

// randomAlphabetChar picks one character uniformly at random. Bytes past the
// last whole multiple of len(alphabet) are rejected so every character is
// equally likely.
func randomAlphabetChar(alphabet string) (byte, error) {
	limit := byte(256 - (256 % len(alphabet)))
	buf := make([]byte, 1)

	for {
		if _, err := rand.Read(buf); err != nil {
			return 0, err
		}
		if buf[0] < limit {
			return alphabet[int(buf[0])%len(alphabet)], nil
		}
	}
}

// normalizeRecoveryCode canonicalizes a code for hashing: uppercase, with the
// dash and any stray spaces removed, so a code entered by hand still matches.
func normalizeRecoveryCode(raw string) string {
	var b strings.Builder

	for _, r := range strings.ToUpper(raw) {
		switch {
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		}
	}

	return b.String()
}
