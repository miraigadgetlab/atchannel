package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"atchannel-backend/internal/service"

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
}

func NewAuthHandler(ts *service.TokenService, us *service.UserService, ss *service.SessionService) *AuthHandler {
	return &AuthHandler{
		tokenService:   ts,
		userService:    us,
		sessionService: ss,
	}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req RegisterRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request payload format",
		})
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	switch {
	case req.Name == "" || req.Email == "" || req.Password == "":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Name, email and password are required",
		})
	case !strings.Contains(req.Email, "@"):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid email address",
		})
	case len(req.Password) < 8:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Password must be at least 8 characters long",
		})
	}

	user, err := h.userService.Create(c.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":   "Conflict",
				"message": "An account with this email already exists",
			})
		case errors.Is(err, service.ErrNameTaken):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":   "Conflict",
				"message": "This name is already taken",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to create account",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req LoginRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request payload format",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Email and password are required",
		})
	}

	user, err := h.userService.Authenticate(c.Context(), req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Invalid email or password",
		})
	}

	userIDStr := strconv.FormatUint(uint64(user.ID), 10)

	tokens, err := h.tokenService.GenerateTokenPair(userIDStr, user.Email, user.Roles)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to generate security tokens",
		})
	}

	// Track the refresh token so it can be revoked later.
	if err := h.sessionService.Record(c.Context(), user.ID, tokens.RefreshToken, time.Unix(tokens.RefreshExpiresAt, 0)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create session",
		})
	}

	return c.JSON(tokens)
}

func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req RefreshRequest
	if err := c.Bind().Body(&req); err != nil || req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Refresh token is required",
		})
	}

	claims, err := h.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired refresh token",
		})
	}

	// The token must still be an active session: not revoked, not expired.
	valid, err := h.sessionService.IsValid(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}
	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Refresh token is no longer valid",
		})
	}

	ownerID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired refresh token",
		})
	}

	tokens, err := h.tokenService.GenerateTokenPair(claims.UserID, claims.Email, claims.Roles)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to generate security tokens",
		})
	}

	// Rotation: the presented token dies, the new one takes its place.
	if _, err := h.sessionService.Revoke(c.Context(), req.RefreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}
	if err := h.sessionService.Record(c.Context(), uint(ownerID), tokens.RefreshToken, time.Unix(tokens.RefreshExpiresAt, 0)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "refresh_token is required",
		})
	}

	revoked, err := h.sessionService.Revoke(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}

	return c.JSON(fiber.Map{"revoked": revoked})
}

// LogoutAll kills every session of the caller ("log out everywhere").
func (h *AuthHandler) LogoutAll(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
	}

	revoked, err := h.sessionService.RevokeAllForUser(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}

	return c.JSON(fiber.Map{"revoked_count": revoked})
}
