package handler

import (
	"errors"
	"strings"

	"atchannel-backend/internal/service"
	"atchannel-backend/internal/validation"

	"github.com/gofiber/fiber/v3"
)

// AccountHandler exposes the account flows that do not require the caller to
// be signed in: resetting a password with a single-use recovery code.
type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(as *service.AccountService) *AccountHandler {
	return &AccountHandler{accountService: as}
}

// ResetPasswordWithCode sets a new password from one of the account's
// single-use recovery codes and signs the user out of every existing session.
// The account is named by its display name; a name that does not exist and a
// wrong code deliberately return the same error, so the endpoint cannot be
// used to enumerate accounts.
func (h *AccountHandler) ResetPasswordWithCode(c fiber.Ctx) error {
	var req struct {
		Name        string `json:"name"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	name := strings.TrimSpace(req.Name)

	switch {
	case name == "" || strings.TrimSpace(req.Code) == "":
		return replyBadRequest(c, "name and code are required")
	case validation.NameError(name) != "":
		return replyBadRequest(c, validation.NameError(name))
	case validation.PasswordError(req.NewPassword) != "":
		return replyBadRequest(c, validation.PasswordError(req.NewPassword))
	}

	if err := h.accountService.ResetPasswordWithCode(c.Context(), name, req.Code, req.NewPassword); err != nil {
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
