package handler

import (
	"strconv"
	"strings"

	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

type SetRolesRequest struct {
	Roles []string `json:"roles"`
}

type AdminHandler struct {
	userService    *service.UserService
	statsService   *service.StatsService
	sessionService *service.SessionService
}

func NewAdminHandler(us *service.UserService, ss *service.StatsService, sessions *service.SessionService) *AdminHandler {
	return &AdminHandler{userService: us, statsService: ss, sessionService: sessions}
}

func (h *AdminHandler) ListUsers(c fiber.Ctx) error {
	limit, offset, err := parsePagination(c)
	if err != nil {
		return replyBadRequest(c, err.Error())
	}

	users, total, err := h.userService.List(c.Context(), limit, offset)
	if err != nil {
		return replyInternal(c, err, "Failed to fetch users")
	}

	return c.JSON(fiber.Map{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"items":  users,
	})
}

func (h *AdminHandler) Stats(c fiber.Ctx) error {
	stats, err := h.statsService.Get(c.Context())
	if err != nil {
		return replyInternal(c, err, "Failed to fetch stats")
	}

	return c.JSON(stats)
}

func (h *AdminHandler) SetRoles(c fiber.Ctx) error {
	targetID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return replyBadRequest(c, "User id must be a number")
	}

	var req SetRolesRequest
	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	if !service.ValidRoles(req.Roles) {
		return replyBadRequest(c, "Roles must be a non-empty subset of: "+strings.Join(service.AllowedRoles, ", "))
	}

	callerID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	if uint(targetID) == callerID {
		return replyBadRequest(c, "You cannot change your own roles")
	}

	user, err := h.userService.SetRoles(c.Context(), uint(targetID), req.Roles)
	if err != nil {
		return svcErr(c, err, "Failed to update roles")
	}

	// Roles live inside the JWT, so an old access token keeps its previous
	// permissions until it expires. Killing every session forces the change
	// to take effect now rather than in up to 15 minutes.
	revoked, err := h.sessionService.RevokeAllForUser(c.Context(), uint(targetID), "roles-changed")
	if err != nil {
		// Not fatal: the role change itself landed. Report it so an admin
		// knows the target may still be signed in briefly.
		return c.JSON(fiber.Map{"user": user, "sessions_revoked": 0, "warning": "roles changed but sessions could not be revoked"})
	}

	return c.JSON(fiber.Map{"user": user, "sessions_revoked": revoked})
}

func (h *AdminHandler) DeleteUser(c fiber.Ctx) error {
	targetID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return replyBadRequest(c, "User id must be a number")
	}

	callerID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	if uint(targetID) == callerID {
		return replyBadRequest(c, "You cannot delete your own account")
	}

	if err := h.userService.Delete(c.Context(), uint(targetID)); err != nil {
		return svcErr(c, err, "Failed to delete user")
	}

	return c.JSON(fiber.Map{"deleted": true})
}
