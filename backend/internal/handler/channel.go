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

// UpdateChannelRequest mirrors the partial-update style of posts: an
// omitted field is left alone, an explicit "" clears it (description).
type UpdateChannelRequest struct {
	Name        *string `json:"name"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

// ChannelDetailResponse is the channel view with a paginated slice of its posts.
type ChannelDetailResponse struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	CreatedBy   *uint         `json:"created_by,omitempty"`
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
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
	}

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

	channel, err := h.channelService.Create(c.Context(), userID, req.Name, req.Title, strings.TrimSpace(req.Description))
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
	query, err := searchQuery(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": err.Error(),
		})
	}

	channels, err := h.channelService.List(c.Context(), service.ChannelFilter{Query: query})
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
		CreatedBy:   channel.CreatedBy,
		CreatedAt:   channel.CreatedAt,
		PostCount:   postCount,
		Posts:       posts,
	})
}

// Update edits a channel: admins may edit any channel, other users only
// the ones they opened (see ChannelService.Update for the check).
func (h *ChannelHandler) Update(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Channel id must be a number",
		})
	}

	var req UpdateChannelRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request payload format",
		})
	}

	if req.Name == nil && req.Title == nil && req.Description == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Nothing to update: provide name, title and/or description",
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
		*req.Name = trimmed
	}
	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Title cannot be empty",
			})
		}
		*req.Title = trimmed
	}

	channel, err := h.channelService.Update(c.Context(), uint(id), userID, currentIsAdmin(c), service.UpdateChannelInput{
		Name:        req.Name,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrChannelNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Channel not found",
			})
		case errors.Is(err, service.ErrForbidden):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Forbidden",
				"message": "Only the channel creator or an admin can edit this channel",
			})
		case errors.Is(err, service.ErrChannelNameTaken):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":   "Conflict",
				"message": "A channel with this name already exists",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to update channel",
			})
		}
	}

	return c.JSON(channel)
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
