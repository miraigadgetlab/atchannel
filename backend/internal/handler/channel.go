package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"atchannel-backend/internal/models"
	"atchannel-backend/internal/service"
	"atchannel-backend/internal/validation"

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
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	var req CreateChannelRequest

	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if msg := validation.ChannelNameError(req.Name); msg != "" {
		return replyBadRequest(c, msg)
	}
	if msg := validation.TooLong("title", req.Title, validation.MaxChannelTitle); msg != "" {
		return replyBadRequest(c, msg)
	}
	if req.Title == "" {
		return replyBadRequest(c, "Name and title are required")
	}
	if msg := validation.TooLong("description", req.Description, validation.MaxChannelDescription); msg != "" {
		return replyBadRequest(c, msg)
	}

	channel, err := h.channelService.Create(c.Context(), userID, req.Name, req.Title, req.Description)
	if err != nil {
		return svcErr(c, err, "Failed to create channel")
	}

	return c.Status(fiber.StatusCreated).JSON(channel)
}

func (h *ChannelHandler) List(c fiber.Ctx) error {
	query, err := searchQuery(c)
	if err != nil {
		return replyBadRequest(c, err.Error())
	}

	limit, offset, err := parsePagination(c)
	if err != nil {
		return replyBadRequest(c, err.Error())
	}

	filter := service.ChannelFilter{Query: query, Limit: limit, Offset: offset}

	channels, err := h.channelService.List(c.Context(), filter)
	if err != nil {
		return replyInternal(c, err, "Failed to fetch channels")
	}

	// Total is counted without the paging clause so the client can render
	// a page count even when this page is the last one.
	total, err := h.channelService.Count(c.Context(), service.ChannelFilter{Query: query})
	if err != nil {
		return replyInternal(c, err, "Failed to count channels")
	}

	return c.JSON(fiber.Map{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"items":  channels,
	})
}

func (h *ChannelHandler) GetByID(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return replyBadRequest(c, "Channel id must be a number")
	}

	channel, err := h.channelService.GetByID(c.Context(), uint(id))
	if err != nil {
		return svcErr(c, err, "Failed to fetch channel")
	}

	limit, offset, err := parsePagination(c)
	if err != nil {
		return replyBadRequest(c, err.Error())
	}

	postCount, err := h.channelService.CountPosts(c.Context(), uint(id))
	if err != nil {
		return replyInternal(c, err, "Failed to count posts")
	}

	channelID := uint(id)
	posts, err := h.postService.List(c.Context(), service.PostFilter{
		ChannelID: &channelID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return replyInternal(c, err, "Failed to fetch posts")
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
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return replyBadRequest(c, "Channel id must be a number")
	}

	var req UpdateChannelRequest

	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	if req.Name == nil && req.Title == nil && req.Description == nil {
		return replyBadRequest(c, "Nothing to update: provide name, title and/or description")
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if msg := validation.ChannelNameError(trimmed); msg != "" {
			return replyBadRequest(c, msg)
		}
		*req.Name = trimmed
	}
	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			return replyBadRequest(c, "Title cannot be empty")
		}
		if msg := validation.TooLong("title", trimmed, validation.MaxChannelTitle); msg != "" {
			return replyBadRequest(c, msg)
		}
		*req.Title = trimmed
	}
	if req.Description != nil {
		if msg := validation.TooLong("description", *req.Description, validation.MaxChannelDescription); msg != "" {
			return replyBadRequest(c, msg)
		}
	}

	channel, err := h.channelService.Update(c.Context(), uint(id), userID, currentIsAdmin(c), service.UpdateChannelInput{
		Name:        req.Name,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return channelErrorResponse(c, err, "Failed to update channel", "edit")
	}

	return c.JSON(channel)
}

// Delete lets the creator (or an admin) take a channel down.
func (h *ChannelHandler) Delete(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return replyBadRequest(c, "Channel id must be a number")
	}

	if err := h.channelService.Delete(c.Context(), uint(id), userID, currentIsAdmin(c)); err != nil {
		return channelErrorResponse(c, err, "Failed to delete channel", "delete")
	}

	return c.JSON(fiber.Map{"deleted": true})
}

// channelErrorResponse maps the service errors shared by update/delete to
// HTTP responses. It is svcErr plus the channel-specific "not your channel"
// wording, which names the action the caller was attempting.
func channelErrorResponse(c fiber.Ctx, err error, internalMessage, action string) error {
	if errors.Is(err, service.ErrForbidden) {
		return replyErr(c, fiber.StatusForbidden, "Only the channel creator or an admin can "+action+" this channel")
	}
	return svcErr(c, err, internalMessage)
}
