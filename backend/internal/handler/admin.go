package handler

import (
	"errors"
	"strconv"
	"strings"

	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

type SetRolesRequest struct {
	Roles []string `json:"roles"`
}

type AdminHandler struct {
	userService  *service.UserService
	statsService *service.StatsService
}

func NewAdminHandler(us *service.UserService, ss *service.StatsService) *AdminHandler {
	return &AdminHandler{userService: us, statsService: ss}
}

func (h *AdminHandler) ListUsers(c fiber.Ctx) error {
	users, err := h.userService.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to fetch users",
		})
	}

	return c.JSON(users)
}

func (h *AdminHandler) Stats(c fiber.Ctx) error {
	stats, err := h.statsService.Get(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to fetch stats",
		})
	}

	return c.JSON(stats)
}

func (h *AdminHandler) SetRoles(c fiber.Ctx) error {
	targetID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "User id must be a number",
		})
	}

	var req SetRolesRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request payload format",
		})
	}

	if !service.ValidRoles(req.Roles) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Roles must be a non-empty subset of: " + joinRoles(service.AllowedRoles),
		})
	}

	callerID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
	}

	if uint(targetID) == callerID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "You cannot change your own roles",
		})
	}

	user, err := h.userService.SetRoles(c.Context(), uint(targetID), req.Roles)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update roles",
		})
	}

	return c.JSON(user)
}

func (h *AdminHandler) DeleteUser(c fiber.Ctx) error {
	targetID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "User id must be a number",
		})
	}

	callerID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
	}

	if uint(targetID) == callerID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "You cannot delete your own account",
		})
	}

	if err := h.userService.Delete(c.Context(), uint(targetID)); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete user",
		})
	}

	return c.JSON(fiber.Map{"deleted": true})
}

func joinRoles(roles []string) string {
	return strings.Join(roles, ", ")
}
