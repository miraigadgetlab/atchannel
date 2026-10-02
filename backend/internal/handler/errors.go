package handler

import (
	"errors"
	"log"
	"net/http"

	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

// replyErr writes the one error envelope every endpoint uses: "error" is the
// HTTP reason phrase, "message" is the human detail.
func replyErr(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error":   http.StatusText(status),
		"message": message,
	})
}

// replyBadRequest is replyErr for the very common 400.
func replyBadRequest(c fiber.Ctx, message string) error {
	return replyErr(c, fiber.StatusBadRequest, message)
}

// replyInternal is replyErr for the very common 500. fallback is what the
// client is told; the underlying error is logged so it is not lost.
func replyInternal(c fiber.Ctx, err error, fallback string) error {
	log.Printf("handler error (%s): %v", fallback, err)
	return replyErr(c, fiber.StatusInternalServerError, fallback)
}

// logSessionFailure records a problem that does not change the outcome the
// caller is told about (the main action already happened), so the failure is
// not silently dropped.
func logSessionFailure(c fiber.Ctx, action string, err error) {
	log.Printf("session cleanup after %s failed: %v", action, err)
}

// svcErr maps the shared service errors to a response and logs anything it
// does not recognise as a 500 with the supplied client-facing fallback.
//
// Keeping the mapping here means a new sentinel error can never be
// accidentally reported as a 500 with a leaked internal message.
func svcErr(c fiber.Ctx, err error, fallback string) error {
	switch {
	case errors.Is(err, service.ErrPostNotFound):
		return replyErr(c, fiber.StatusNotFound, "Post does not exist")
	case errors.Is(err, service.ErrChannelNotFound):
		return replyErr(c, fiber.StatusNotFound, "Channel not found")
	case errors.Is(err, service.ErrCommentNotFound):
		return replyErr(c, fiber.StatusNotFound, "Comment not found")
	case errors.Is(err, service.ErrUserNotFound):
		return replyErr(c, fiber.StatusNotFound, "User not found")
	case errors.Is(err, service.ErrChannelNameTaken):
		return replyErr(c, fiber.StatusConflict, "A channel with this name already exists")
	case errors.Is(err, service.ErrEmailTaken):
		return replyErr(c, fiber.StatusConflict, "Email is already registered")
	case errors.Is(err, service.ErrNameTaken):
		return replyErr(c, fiber.StatusConflict, "This name is already taken")
	case errors.Is(err, service.ErrInvalidCredentials):
		return replyErr(c, fiber.StatusUnauthorized, "Invalid email or password")
	case errors.Is(err, service.ErrCurrentPassword):
		return replyErr(c, fiber.StatusBadRequest, "Current password is incorrect")
	case errors.Is(err, service.ErrInvalidRoles):
		return replyErr(c, fiber.StatusBadRequest, "Roles must be a non-empty subset of: user, admin")
	case errors.Is(err, service.ErrForbidden):
		// Handlers that need bespoke wording check for this first; the rest
		// must still land on 403 rather than a misleading 500.
		return replyErr(c, fiber.StatusForbidden, "You do not have permission to perform this action")
	default:
		return replyInternal(c, err, fallback)
	}
}
