package handler

import (
	"errors"
	"strings"

	"atchannel-backend/internal/service"
	"atchannel-backend/internal/validation"

	"github.com/gofiber/fiber/v3"
)

// AccountHandler exposes the account flows that do not require the caller to
// be the owner only via email: emailed password reset, single-use recovery
// codes and the account's private admin note.
type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(as *service.AccountService) *AccountHandler {
	return &AccountHandler{accountService: as}
}

// ForgotPassword always answers the same way, whether or not the address is
// registered, so it cannot be used to enumerate accounts.
func (h *AccountHandler) ForgotPassword(c fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	if email := strings.ToLower(strings.TrimSpace(req.Email)); validation.EmailError(email) == "" {
		h.accountService.RequestPasswordReset(c.Context(), email)
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// ResetPassword sets a new password from a reset token and signs the user out
// of every existing session.
func (h *AccountHandler) ResetPassword(c fiber.Ctx) error {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	switch {
	case strings.TrimSpace(req.Token) == "":
		return replyBadRequest(c, "token is required")
	case validation.PasswordError(req.NewPassword) != "":
		return replyBadRequest(c, validation.PasswordError(req.NewPassword))
	}

	if err := h.accountService.ResetPassword(c.Context(), req.Token, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidEmailToken) {
			return replyErr(c, fiber.StatusBadRequest, "This reset link is invalid or has expired")
		}
		return svcErr(c, err, "Failed to reset password")
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// ResetPasswordWithCode resets a password from a single-use recovery code
// instead of an emailed link — the path for accounts without a working mailbox.
func (h *AccountHandler) ResetPasswordWithCode(c fiber.Ctx) error {
	var req struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	switch {
	case email == "" || strings.TrimSpace(req.Code) == "":
		return replyBadRequest(c, "email and code are required")
	case validation.EmailError(email) != "":
		return replyBadRequest(c, "invalid email address")
	case validation.PasswordError(req.NewPassword) != "":
		return replyBadRequest(c, validation.PasswordError(req.NewPassword))
	}

	if err := h.accountService.ResetPasswordWithCode(c.Context(), email, req.Code, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidRecoveryCode) {
			return replyErr(c, fiber.StatusBadRequest, "This recovery code is invalid or has already been used")
		}
		return svcErr(c, err, "Failed to reset password")
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// RecoveryCodes issues a fresh batch of single-use recovery codes for the
// caller. The plaintext codes are returned exactly once; only hashes are kept,
// and generating a new batch retires every previous code.
func (h *AccountHandler) RecoveryCodes(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	codes, err := h.accountService.GenerateRecoveryCodes(c.Context(), userID)
	if err != nil {
		return replyInternal(c, err, "Failed to generate recovery codes")
	}

	return c.JSON(fiber.Map{"codes": codes})
}

// RecoveryCodeStatus reports how many unused recovery codes the caller has left.
func (h *AccountHandler) RecoveryCodeStatus(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	remaining, err := h.accountService.RecoveryCodeStatus(c.Context(), userID)
	if err != nil {
		return replyInternal(c, err, "Failed to fetch recovery codes")
	}

	return c.JSON(fiber.Map{"remaining": remaining})
}
