package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"atchannel-backend/internal/mailer"
	"atchannel-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Email token kinds. Keeping them distinct means a reset link cannot be
// replayed as a recovery code and vice versa. Recovery codes share the same
// storage as emailed tokens but identify the account on their own.
const (
	TokenKindReset    = "reset"
	TokenKindRecovery = "recovery"
)

// RecoveryCodeCount is how many single-use codes one batch contains.
const RecoveryCodeCount = 10

// recoveryAlphabet omits the characters that are easy to confuse when a code
// is written down or read back (0/O, 1/I/L).
const recoveryAlphabet = "23456789ABCDEFGHJKMNPQRSTVWXYZ"

var (
	ErrInvalidEmailToken   = errors.New("email token is invalid or has expired")
	ErrInvalidRecoveryCode = errors.New("recovery code is invalid or has already been used")
)

// AccountService owns the account's credential flows: password reset (by
// emailed link or single-use recovery code) and the issuing and consumption of
// those tokens. Only hashes are ever stored.
type AccountService struct {
	db       *gorm.DB
	users    *UserService
	sessions *SessionService
	mailer   mailer.Mailer
	baseURL  string
	resetTTL time.Duration
}

func NewAccountService(
	db *gorm.DB,
	users *UserService,
	sessions *SessionService,
	m mailer.Mailer,
	baseURL string,
	resetTTL time.Duration,
) *AccountService {
	return &AccountService{
		db:       db,
		users:    users,
		sessions: sessions,
		mailer:   m,
		baseURL:  baseURL,
		resetTTL: resetTTL,
	}
}

// RequestPasswordReset emails a reset link when the address belongs to an
// account. It reports nothing either way, so the endpoint cannot be used to
// enumerate registered emails.
func (s *AccountService) RequestPasswordReset(ctx context.Context, email string) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		log.Printf("password reset requested for unknown email %q", email)
		return
	}

	raw, err := s.issueToken(ctx, user.ID, TokenKindReset, s.resetTTL)
	if err != nil {
		log.Printf("could not issue reset token for user %d: %v", user.ID, err)
		return
	}

	s.deliver(user.Email, "Reset your password", fmt.Sprintf(
		"Someone (hopefully you) asked to reset the password for your atchannel account.\n\nSet a new one by visiting:\n%s/reset-password?token=%s\n\nThis link expires in %s. If you did not ask, you can ignore this email.\n",
		s.baseURL, raw, humanize(s.resetTTL),
	))
}

// ResetPassword consumes a reset token, stores the new password and signs the
// user out everywhere: changing the password is how you evict whoever knew the
// old one.
func (s *AccountService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	userID, err := s.consumeToken(ctx, rawToken, TokenKindReset)
	if err != nil {
		return err
	}

	return s.setPassword(ctx, userID, newPassword)
}

// ResetPasswordWithCode is the email-free recovery path: a single-use recovery
// code stands in for the emailed link. The address is only used to look the
// account up, so a wrong code can never reset somebody else's password.
func (s *AccountService) ResetPasswordWithCode(ctx context.Context, email, code, newPassword string) error {
	user, err := s.users.GetByEmail(ctx, email)
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
	rows := make([]models.EmailToken, 0, RecoveryCodeCount)

	for i := 0; i < RecoveryCodeCount; i++ {
		code, err := randomRecoveryCode()
		if err != nil {
			return nil, err
		}
		codes = append(codes, code)
		rows = append(rows, models.EmailToken{
			UserID:    userID,
			TokenHash: HashToken(normalizeRecoveryCode(code)),
			Kind:      TokenKindRecovery,
			// Recovery codes do not meaningfully expire; the far-future
			// value only satisfies the NOT NULL column.
			ExpiresAt: time.Now().AddDate(100, 0, 0),
		})
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Regenerating invalidates the old batch entirely.
		if err := tx.Where("user_id = ? AND kind = ?", userID, TokenKindRecovery).
			Delete(&models.EmailToken{}).Error; err != nil {
			return err
		}
		return tx.Create(&rows).Error
	})
	if err != nil {
		return nil, err
	}

	return codes, nil
}

// RecoveryCodeStatus reports how many unused recovery codes remain, so the
// settings page can nudge the user before they run out.
func (s *AccountService) RecoveryCodeStatus(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.EmailToken{}).
		Where("user_id = ? AND kind = ? AND used_at IS NULL", userID, TokenKindRecovery).
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

// issueToken invalidates any outstanding tokens of the same kind for the user
// (only the newest link should work) and returns the raw, un-hashed token.
func (s *AccountService) issueToken(ctx context.Context, userID uint, kind string, ttl time.Duration) (string, error) {
	raw, err := randomToken()
	if err != nil {
		return "", err
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND kind = ? AND used_at IS NULL", userID, kind).
			Delete(&models.EmailToken{}).Error; err != nil {
			return err
		}
		return tx.Create(&models.EmailToken{
			UserID:    userID,
			TokenHash: HashToken(raw),
			Kind:      kind,
			ExpiresAt: time.Now().Add(ttl),
		}).Error
	})
	if err != nil {
		return "", err
	}

	return raw, nil
}

// consumeToken validates and burns a token in one transaction, returning the
// account it belonged to.
func (s *AccountService) consumeToken(ctx context.Context, rawToken, kind string) (uint, error) {
	var userID uint
	var outcome error

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row models.EmailToken

		// FOR UPDATE so two concurrent redemptions of the same link cannot
		// both succeed.
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND kind = ? AND used_at IS NULL", HashToken(rawToken), kind).
			First(&row).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				outcome = ErrInvalidEmailToken
				return nil
			}
			return err
		}

		if time.Now().After(row.ExpiresAt) {
			outcome = ErrInvalidEmailToken
			return nil
		}

		if err := tx.Model(&row).Update("used_at", time.Now()).Error; err != nil {
			return err
		}
		userID = row.UserID
		return nil
	})
	if err != nil {
		return 0, err
	}
	if outcome != nil {
		return 0, outcome
	}

	return userID, nil
}

// consumeRecoveryCode validates and burns one single-use code for the account.
// Unlike emailed tokens, recovery codes never expire: they stay valid until
// they are used or the owner regenerates the batch.
func (s *AccountService) consumeRecoveryCode(ctx context.Context, userID uint, code string) error {
	var outcome error

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row models.EmailToken

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND token_hash = ? AND kind = ? AND used_at IS NULL",
				userID, HashToken(normalizeRecoveryCode(code)), TokenKindRecovery).
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

func (s *AccountService) deliver(to, subject, body string) {
	go func() {
		if err := s.mailer.Send(mailer.Message{To: to, Subject: subject, Body: body}); err != nil {
			log.Printf("mailer: sending to %s failed: %v", to, err)
		}
	}()
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
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

func humanize(d time.Duration) string {
	if d >= time.Hour {
		return fmt.Sprintf("%.0f hours", d.Hours())
	}
	return fmt.Sprintf("%.0f minutes", d.Minutes())
}
