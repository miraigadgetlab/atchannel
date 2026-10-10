package handler

import (
	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

// NotificationHandler serves the caller's activity feed.
type NotificationHandler struct {
	notifications *service.NotificationService
}

func NewNotificationHandler(ns *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifications: ns}
}

func (h *NotificationHandler) List(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	limit, offset, err := parsePagination(c)
	if err != nil {
		return replyBadRequest(c, err.Error())
	}

	items, total, err := h.notifications.List(c.Context(), userID, limit, offset)
	if err != nil {
		return replyInternal(c, err, "Failed to fetch notifications")
	}

	return c.JSON(fiber.Map{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"items":  toNotificationResponses(items),
	})
}

func (h *NotificationHandler) UnreadCount(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	count, err := h.notifications.UnreadCount(c.Context(), userID)
	if err != nil {
		return replyInternal(c, err, "Failed to count notifications")
	}

	return c.JSON(fiber.Map{"unread": count})
}

func (h *NotificationHandler) MarkAllRead(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	updated, err := h.notifications.MarkAllRead(c.Context(), userID)
	if err != nil {
		return replyInternal(c, err, "Failed to mark notifications read")
	}

	return c.JSON(fiber.Map{"updated": updated})
}
