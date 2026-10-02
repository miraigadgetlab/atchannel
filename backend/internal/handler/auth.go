package handler

import (
	"errors"
	"strconv"
	"strings"

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
	tokenService *service.TokenService
	userService  *service.UserService
}

func NewAuthHandler(ts *service.TokenService, us *service.UserService) *AuthHandler {
	return &AuthHandler{
		tokenService: ts,
		userService:  us,
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

	tokens, err := h.tokenService.RefreshTokens(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired refresh token",
		})
	}

	return c.JSON(tokens)
}
