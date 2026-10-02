package handler

import (
	"errors"
	"strings"

	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

const (
	maxAboutMeLength   = 1000
	maxAvatarURLLength = 500
	minPasswordLength  = 8
)

type UpdateProfileRequest struct {
	Name      *string `json:"name"`
	Email     *string `json:"email"`
	AboutMe   *string `json:"about_me"`
	AvatarURL *string `json:"avatar_url"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type ProfileHandler struct {
	userService *service.UserService
}

func NewProfileHandler(us *service.UserService) *ProfileHandler {
	return &ProfileHandler{userService: us}
}

// Me returns the caller's own account record (password is never serialized).
func (h *ProfileHandler) Me(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
	}

	user, err := h.userService.GetByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Account no longer exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to fetch profile",
		})
	}

	return c.JSON(user)
}

func (h *ProfileHandler) Update(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
	}

	var req UpdateProfileRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request payload format",
		})
	}

	if req.Name == nil && req.Email == nil && req.AboutMe == nil && req.AvatarURL == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Nothing to update: provide name, email, about_me and/or avatar_url",
		})
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Name cannot be empty",
			})
		}
		req.Name = &trimmed
	}

	if req.Email != nil {
		trimmed := strings.ToLower(strings.TrimSpace(*req.Email))
		if !strings.Contains(trimmed, "@") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid email address",
			})
		}
		req.Email = &trimmed
	}

	if req.AboutMe != nil && len(*req.AboutMe) > maxAboutMeLength {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "about_me must be at most 1000 characters long",
		})
	}

	if req.AvatarURL != nil {
		trimmed := strings.TrimSpace(*req.AvatarURL)
		if len(trimmed) > maxAvatarURLLength {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "avatar_url must be at most 500 characters long",
			})
		}
		req.AvatarURL = &trimmed
	}

	user, err := h.userService.UpdateProfile(c.Context(), userID, service.UpdateProfileInput{
		Name:      req.Name,
		Email:     req.Email,
		AboutMe:   req.AboutMe,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNameTaken):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":   "Conflict",
				"message": "This name is already taken",
			})
		case errors.Is(err, service.ErrEmailTaken):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":   "Conflict",
				"message": "An account with this email already exists",
			})
		case errors.Is(err, service.ErrUserNotFound):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Account no longer exists",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to update profile",
			})
		}
	}

	return c.JSON(user)
}

func (h *ProfileHandler) ChangePassword(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
	}

	var req ChangePasswordRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request payload format",
		})
	}

	switch {
	case req.CurrentPassword == "" || req.NewPassword == "":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "current_password and new_password are required",
		})
	case len(req.NewPassword) < minPasswordLength:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "New password must be at least 8 characters long",
		})
	case req.NewPassword == req.CurrentPassword:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "New password must differ from the current one",
		})
	}

	if err := h.userService.ChangePassword(c.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, service.ErrCurrentPassword):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Current password is incorrect",
			})
		case errors.Is(err, service.ErrUserNotFound):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Account no longer exists",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to change password",
			})
		}
	}

	return c.JSON(fiber.Map{"changed": true})
}
