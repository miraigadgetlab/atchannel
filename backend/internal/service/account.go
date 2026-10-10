package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"atchannel-backend/internal/mailer"
	"atchannel-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Email token kinds. Keeping them distinct means a reset link cannot be
// replayed as a verification and vice versa.
const (
	TokenKindVerify = "verify"
	TokenKindReset  = "reset"
)

var (
	ErrInvalidEmailToken    = errors.New("email token is invalid or has expired")
	ErrEmailAlreadyVerified = errors.New("email is already verified")
)

// AccountService owns the email-driven account flows: address verification and
// password reset. Tokens are random, single-use and stored only as hashes.
type AccountService struct {
	db        *gorm.DB
	users     *UserService
	sessions  *SessionService
	mailer    mailer.Mailer
	baseURL   string
	verifyTTL time.Duration
	resetTTL  time.Duration
}

func NewAccountService(
	db *gorm.DB,
	users *UserService,
	sessions *SessionService,
	m mailer.Mailer,
	baseURL string,
	verifyTTL, resetTTL time.Duration,
) *AccountService {
	return &AccountService{
		db:        db,
		users:     users,
		sessions:  sessions,
		mailer:    m,
		baseURL:   baseURL,
		verifyTTL: verifyTTL,
		resetTTL:  resetTTL,
	}
}

// SendVerification issues a fresh verification token and emails the link. The
// caller gets an error only when the token cannot be stored; a failure to
// deliver is logged, not surfaced, because the account already exists and the
// user can ask for another link.
func (s *AccountService) SendVerification(ctx context.Context, user *models.User) error {
	raw, err := s.issueToken(ctx, user.ID, TokenKindVerify, s.verifyTTL)
	if err != nil {
		return err
	}

	s.deliver(user.Email, "Verify your email address", fmt.Sprintf(
		"Welcome to atchannel!\n\nConfirm this email address by visiting:\n%s/verify-email?token=%s\n\nThis link expires in %s.\n",
		s.baseURL, raw, humanize(s.verifyTTL),
	))
	return nil
}

// ResendVerification re-sends the confirmation link to an existing account.
func (s *AccountService) ResendVerification(ctx context.Context, userID uint) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.EmailVerifiedAt != nil {
		return ErrEmailAlreadyVerified
	}
	return s.SendVerification(ctx, user)
}

// VerifyEmail consumes a verification token and marks the account verified.
func (s *AccountService) VerifyEmail(ctx context.Context, rawToken string) error {
	userID, err := s.consumeToken(ctx, rawToken, TokenKindVerify)
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("email_verified_at", time.Now()).Error
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

func humanize(d time.Duration) string {
	if d >= time.Hour {
		return fmt.Sprintf("%.0f hours", d.Hours())
	}
	return fmt.Sprintf("%.0f minutes", d.Minutes())
}
