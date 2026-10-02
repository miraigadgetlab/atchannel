package handler

import (
	"errors"
	"strconv"
	"strings"

	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

type CreatePostRequest struct {
	ChannelID uint   `json:"channel_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
}

type UpdatePostRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(ps *service.PostService) *PostHandler {
	return &PostHandler{postService: ps}
}

func (h *PostHandler) Create(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
	}

	var req CreatePostRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request payload format",
		})
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)

	switch {
	case req.ChannelID == 0:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "channel_id is required",
		})
	case req.Title == "" || req.Content == "":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Title and content are required",
		})
	}

	post, err := h.postService.Create(c.Context(), uint(userID), req.ChannelID, req.Title, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrChannelNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Channel does not exist",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create post",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(post)
}

func (h *PostHandler) List(c fiber.Ctx) error {
	filter := service.PostFilter{}

	if raw := c.Query("channel_id"); raw != "" {
		channelID, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "channel_id must be a number",
			})
		}
		id := uint(channelID)
		filter.ChannelID = &id
	}

	if raw := c.Query("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "limit must be a number between 1 and 100",
			})
		}
		filter.Limit = limit
	} else {
		filter.Limit = 20
	}

	if raw := c.Query("offset"); raw != "" {
		offset, err := strconv.Atoi(raw)
		if err != nil || offset < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "offset must be a non-negative number",
			})
		}
		filter.Offset = offset
	}

	posts, err := h.postService.List(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to fetch posts",
		})
	}

	return c.JSON(posts)
}

func (h *PostHandler) GetByID(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Post id must be a number",
		})
	}

	post, err := h.postService.GetByID(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Post not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to fetch post",
		})
	}

	return c.JSON(post)
}

func (h *PostHandler) Update(c fiber.Ctx) error {
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
			"message": "Post id must be a number",
		})
	}

	var req UpdatePostRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request payload format",
		})
	}

	if req.Title == nil && req.Content == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Nothing to update: provide title and/or content",
		})
	}

	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Title cannot be empty",
			})
		}
		req.Title = &trimmed
	}

	if req.Content != nil {
		trimmed := strings.TrimSpace(*req.Content)
		if trimmed == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Content cannot be empty",
			})
		}
		req.Content = &trimmed
	}

	post, err := h.postService.Update(c.Context(), uint(id), userID, service.UpdatePostInput{
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		return postErrorResponse(c, err, "Failed to update post")
	}

	return c.JSON(post)
}

func (h *PostHandler) Delete(c fiber.Ctx) error {
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
			"message": "Post id must be a number",
		})
	}

	if err := h.postService.Delete(c.Context(), uint(id), userID); err != nil {
		return postErrorResponse(c, err, "Failed to delete post")
	}

	return c.JSON(fiber.Map{"deleted": true})
}

// postErrorResponse maps service errors shared by update/delete to HTTP responses.
func postErrorResponse(c fiber.Ctx, err error, internalMessage string) error {
	switch {
	case errors.Is(err, service.ErrPostNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "Post not found",
		})
	case errors.Is(err, service.ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   "Forbidden",
			"message": "You can only modify your own posts",
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": internalMessage,
		})
	}
}
