package handler

import (
	"errors"
	"net/url"
	"strings"

	"atchannel-backend/internal/service"
	"atchannel-backend/internal/validation"

	"github.com/gofiber/fiber/v3"
)

const (
	maxAboutMeLength   = validation.MaxAboutMe
	maxAvatarURLLength = validation.MaxAvatarURL
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
	userService    *service.UserService
	sessionService *service.SessionService
}

func NewProfileHandler(us *service.UserService, ss *service.SessionService) *ProfileHandler {
	return &ProfileHandler{userService: us, sessionService: ss}
}

// Me returns the caller's own account record (password is never serialized).
func (h *ProfileHandler) Me(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	user, err := h.userService.GetByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return replyErr(c, fiber.StatusUnauthorized, "Account no longer exists")
		}
		return replyInternal(c, err, "Failed to fetch profile")
	}

	return c.JSON(user)
}

func (h *ProfileHandler) Update(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	var req UpdateProfileRequest

	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	if req.Name == nil && req.Email == nil && req.AboutMe == nil && req.AvatarURL == nil {
		return replyBadRequest(c, "Nothing to update: provide name, email, about_me and/or avatar_url")
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if msg := validation.NameError(trimmed); msg != "" {
			return replyBadRequest(c, msg)
		}
		req.Name = &trimmed
	}

	if req.Email != nil {
		trimmed := strings.ToLower(strings.TrimSpace(*req.Email))
		if msg := validation.EmailError(trimmed); msg != "" {
			return replyBadRequest(c, msg)
		}
		req.Email = &trimmed
	}

	if req.AboutMe != nil && validation.Length(*req.AboutMe) > maxAboutMeLength {
		return replyBadRequest(c, "about_me must be at most 1000 characters long")
	}

	if req.AvatarURL != nil {
		trimmed := strings.TrimSpace(*req.AvatarURL)
		if validation.Length(trimmed) > maxAvatarURLLength {
			return replyBadRequest(c, "avatar_url must be at most 500 characters long")
		}
		// Only http(s) is accepted: a javascript:, data: or file: URL in an
		// avatar field is a stored XSS waiting for a renderer.
		if trimmed != "" && !isSafeURL(trimmed) {
			return replyBadRequest(c, "avatar_url must be an http or https URL")
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
		case errors.Is(err, service.ErrEmailTaken):
			return replyErr(c, fiber.StatusConflict, "An account with this email already exists")
		case errors.Is(err, service.ErrUserNotFound):
			return replyErr(c, fiber.StatusUnauthorized, "Account no longer exists")
		default:
			return svcErr(c, err, "Failed to update profile")
		}
	}

	return c.JSON(user)
}

func (h *ProfileHandler) ChangePassword(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	var req ChangePasswordRequest

	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	switch {
	case req.CurrentPassword == "" || req.NewPassword == "":
		return replyBadRequest(c, "current_password and new_password are required")
	case len(req.NewPassword) < minPasswordLength:
		return replyBadRequest(c, "New password must be at least 8 characters long")
	case req.NewPassword == req.CurrentPassword:
		return replyBadRequest(c, "New password must differ from the current one")
	}

	if err := h.userService.ChangePassword(c.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, service.ErrCurrentPassword):
			// 401 rather than 400: this is a failed credential check, and
			// clients already retry 401s by asking the user to re-authenticate.
			return replyErr(c, fiber.StatusUnauthorized, "Current password is incorrect")
		case errors.Is(err, service.ErrUserNotFound):
			return replyErr(c, fiber.StatusUnauthorized, "Account no longer exists")
		default:
			return replyInternal(c, err, "Failed to change password")
		}
	}

	// A password change signs every existing session out.
	revoked, err := h.sessionService.RevokeAllForUser(c.Context(), userID, "password")
	if err != nil {
		// The password did change, so report the session state honestly
		// rather than pretending the whole operation failed.
		logSessionFailure(c, "password change", err)
		return c.JSON(fiber.Map{"changed": true, "sessions_revoked": 0, "warning": "password changed but sessions could not be revoked"})
	}

	return c.JSON(fiber.Map{"changed": true, "sessions_revoked": revoked})
}

// isSafeURL reports whether raw parses as an absolute http(s) URL. Anything
// else (javascript:, data:, file:, a bare path) is rejected.
func isSafeURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}
