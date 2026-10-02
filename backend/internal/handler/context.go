package handler

import (
	"errors"
	"slices"
	"strconv"

	"atchannel-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

// currentUserID returns the account the auth middleware attached to this
// request. It only fails on handlers mounted behind the auth guard, so a
// failure here means the middleware and the handler disagree — treat it as
// an authentication problem, not a client mistake.
func currentUserID(c fiber.Ctx) (uint, error) {
	claims, ok := c.Locals("user").(*middleware.UserClaims)
	if !ok || claims == nil {
		return 0, errors.New("missing authentication context")
	}

	id, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid user identity in token")
	}

	return uint(id), nil
}

// currentIsAdmin reports whether the caller's token carries the admin role.
func currentIsAdmin(c fiber.Ctx) bool {
	claims, ok := c.Locals("user").(*middleware.UserClaims)
	if !ok || claims == nil {
		return false
	}

	return slices.Contains(claims.Roles, "admin")
}
