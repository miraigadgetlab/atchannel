package handler

import (
	"errors"
	"strings"

	"atchannel-backend/internal/service"
	"atchannel-backend/internal/validation"

	"github.com/gofiber/fiber/v3"
)

// AccountHandler exposes the email-driven account flows: address verification
// and password reset.
type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(as *service.AccountService) *AccountHandler {
	return &AccountHandler{accountService: as}
}

// VerifyEmail confirms an address from a token in the email link.
func (h *AccountHandler) VerifyEmail(c fiber.Ctx) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.Bind().Body(&req); err != nil || strings.TrimSpace(req.Token) == "" {
		return replyBadRequest(c, "token is required")
	}

	if err := h.accountService.VerifyEmail(c.Context(), req.Token); err != nil {
		if errors.Is(err, service.ErrInvalidEmailToken) {
			return replyErr(c, fiber.StatusBadRequest, "This verification link is invalid or has expired")
		}
		return svcErr(c, err, "Failed to verify email")
	}

	return c.JSON(fiber.Map{"verified": true})
}

// ResendVerification re-sends the confirmation link for the caller's account.
func (h *AccountHandler) ResendVerification(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	if err := h.accountService.ResendVerification(c.Context(), userID); err != nil {
		switch {
		case errors.Is(err, service.ErrEmailAlreadyVerified):
			return replyErr(c, fiber.StatusConflict, "Email is already verified")
		default:
			return svcErr(c, err, "Failed to send verification email")
		}
	}

	return c.JSON(fiber.Map{"sent": true})
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
