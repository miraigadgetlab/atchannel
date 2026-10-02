package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"atchannel-backend/internal/models"
	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

type CreateChannelRequest struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ChannelDetailResponse is the channel view with a paginated slice of its posts.
type ChannelDetailResponse struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	CreatedAt   time.Time     `json:"created_at"`
	PostCount   int64         `json:"post_count"`
	Posts       []models.Post `json:"posts"`
}

type ChannelHandler struct {
	channelService *service.ChannelService
	postService    *service.PostService
}

func NewChannelHandler(cs *service.ChannelService, ps *service.PostService) *ChannelHandler {
	return &ChannelHandler{channelService: cs, postService: ps}
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

func (h *ChannelHandler) GetByID(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Channel id must be a number",
		})
	}

	channel, err := h.channelService.GetByID(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrChannelNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Channel not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to fetch channel",
		})
	}

	limit, offset, err := parsePagination(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": err.Error(),
		})
	}

	postCount, err := h.channelService.CountPosts(c.Context(), uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to count posts",
		})
	}

	channelID := uint(id)
	posts, err := h.postService.List(c.Context(), service.PostFilter{
		ChannelID: &channelID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to fetch posts",
		})
	}

	return c.JSON(ChannelDetailResponse{
		ID:          channel.ID,
		Name:        channel.Name,
		Title:       channel.Title,
		Description: channel.Description,
		CreatedAt:   channel.CreatedAt,
		PostCount:   postCount,
		Posts:       posts,
	})
}

// Delete is mounted behind the admin role guard, so there is no ownership check here.
func (h *ChannelHandler) Delete(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Channel id must be a number",
		})
	}

	if err := h.channelService.Delete(c.Context(), uint(id)); err != nil {
		if errors.Is(err, service.ErrChannelNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Channel not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete channel",
		})
	}

	return c.JSON(fiber.Map{"deleted": true})
}
