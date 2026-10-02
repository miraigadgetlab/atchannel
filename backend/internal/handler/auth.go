package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"atchannel-backend/internal/middleware"
	"atchannel-backend/internal/service"
	"atchannel-backend/internal/validation"

	"github.com/gofiber/fiber/v3"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthHandler struct {
	tokenService   *service.TokenService
	userService    *service.UserService
	sessionService *service.SessionService
	accountLimiter *middleware.AccountLimiter
}

func NewAuthHandler(ts *service.TokenService, us *service.UserService, ss *service.SessionService, al *middleware.AccountLimiter) *AuthHandler {
	return &AuthHandler{
		tokenService:   ts,
		userService:    us,
		sessionService: ss,
		accountLimiter: al,
	}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req RegisterRequest

	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	switch {
	case req.Name == "" || req.Email == "" || req.Password == "":
		return replyBadRequest(c, "Name, email and password are required")
	case validation.NameError(req.Name) != "":
		return replyBadRequest(c, validation.NameError(req.Name))
	case validation.EmailError(req.Email) != "":
		return replyBadRequest(c, validation.EmailError(req.Email))
	case len(req.Password) < 8:
		return replyBadRequest(c, "Password must be at least 8 characters long")
	}

	user, err := h.userService.Create(c.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			return replyErr(c, fiber.StatusConflict, "An account with this email already exists")
		}
		return svcErr(c, err, "Failed to create account")
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req LoginRequest

	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	if req.Email == "" || req.Password == "" {
		return replyBadRequest(c, "Email and password are required")
	}

	// Account level lockout: even the right password stays rejected while
	// the window is open, which is the whole point of the brake.
	if wait, blocked := h.accountLimiter.Blocked(req.Email); blocked {
		c.Set(fiber.HeaderRetryAfter, strconv.Itoa(int(wait.Seconds())+1))
		return replyErr(c, fiber.StatusTooManyRequests, "Too many failed attempts for this account, try again later")
	}

	user, err := h.userService.Authenticate(c.Context(), req.Email, req.Password)
	if err != nil {
		h.accountLimiter.Fail(req.Email)
		return replyErr(c, fiber.StatusUnauthorized, "Invalid email or password")
	}

	h.accountLimiter.Clear(req.Email)

	userIDStr := strconv.FormatUint(uint64(user.ID), 10)

	tokens, err := h.tokenService.GenerateTokenPair(userIDStr, user.Email, user.Roles)
	if err != nil {
		return replyInternal(c, err, "Failed to generate security tokens")
	}

	// Track the refresh token so it can be revoked later. A fresh login
	// starts a new family: everything rotated from here descends from it.
	if err := h.sessionService.Record(c.Context(), user.ID, service.NewFamilyID(), tokens.RefreshToken, time.Unix(tokens.RefreshExpiresAt, 0)); err != nil {
		return replyInternal(c, err, "Failed to create session")
	}

	return c.JSON(tokens)
}

func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req RefreshRequest
	if err := c.Bind().Body(&req); err != nil || req.RefreshToken == "" {
		return replyBadRequest(c, "Refresh token is required")
	}

	claims, err := h.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, "Invalid or expired refresh token")
	}

	ownerID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, "Invalid or expired refresh token")
	}

	// The account must still exist: a token outliving its user (deleted
	// account, admin purge) must not resurrect a session.
	if _, err := h.userService.GetByID(c.Context(), uint(ownerID)); err != nil {
		return replyErr(c, fiber.StatusUnauthorized, "Refresh token is no longer valid")
	}

	tokens, err := h.tokenService.GenerateTokenPair(claims.UserID, claims.Email, claims.Roles)
	if err != nil {
		return replyInternal(c, err, "Failed to generate security tokens")
	}

	// Rotation happens in one transaction: the presented token is revoked
	// and the replacement recorded together, and a replayed token burns
	// the whole family instead of quietly minting a fresh one.
	if _, err := h.sessionService.Rotate(
		c.Context(),
		req.RefreshToken,
		tokens.RefreshToken,
		time.Unix(tokens.RefreshExpiresAt, 0),
	); err != nil {
		switch {
		case errors.Is(err, service.ErrTokenReuse):
			return replyErr(c, fiber.StatusUnauthorized, "Session terminated: refresh token was reused, sign in again")
		case errors.Is(err, service.ErrExpiredToken):
			return replyErr(c, fiber.StatusUnauthorized, "Refresh token is no longer valid")
		default:
			return replyInternal(c, err, "Failed to refresh session")
		}
	}

	return c.JSON(tokens)
}

// Logout revokes the presented refresh token. It is intentionally idempotent:
// an unknown or already revoked token just reports revoked:false with a 200.
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	type LogoutRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req LogoutRequest
	if err := c.Bind().Body(&req); err != nil || req.RefreshToken == "" {
		return replyBadRequest(c, "refresh_token is required")
	}

	revoked, err := h.sessionService.Revoke(c.Context(), req.RefreshToken, "logout")
	if err != nil {
		return replyInternal(c, err, "Failed to revoke session")
	}

	return c.JSON(fiber.Map{"revoked": revoked})
}

// LogoutAll kills every session of the caller ("log out everywhere").
func (h *AuthHandler) LogoutAll(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	revoked, err := h.sessionService.RevokeAllForUser(c.Context(), userID, "logout-all")
	if err != nil {
		return replyInternal(c, err, "Failed to revoke sessions")
	}

	return c.JSON(fiber.Map{"revoked_count": revoked})
}
