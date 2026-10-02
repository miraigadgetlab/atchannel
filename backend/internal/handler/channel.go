package handler

import (
	"errors"
	"strings"

	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

type CreateChannelRequest struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ChannelHandler struct {
	channelService *service.ChannelService
}

func NewChannelHandler(cs *service.ChannelService) *ChannelHandler {
	return &ChannelHandler{channelService: cs}
}

func (h *ChannelHandler) Create(c fiber.Ctx) error {
	var req CreateChannelRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request payload format",
		})
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Title = strings.TrimSpace(req.Title)

	if req.Name == "" || req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Name and title are required",
		})
	}

	channel, err := h.channelService.Create(c.Context(), req.Name, req.Title, strings.TrimSpace(req.Description))
	if err != nil {
		if errors.Is(err, service.ErrChannelNameTaken) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":   "Conflict",
				"message": "A channel with this name already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create channel",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(channel)
}

func (h *ChannelHandler) List(c fiber.Ctx) error {
	channels, err := h.channelService.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to fetch channels",
		})
	}

	return c.JSON(channels)
}
