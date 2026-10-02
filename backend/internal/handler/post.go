package handler

import (
	"errors"
	"strconv"
	"strings"

	"atchannel-backend/internal/service"
	"atchannel-backend/internal/validation"

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
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	var req CreatePostRequest

	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)

	switch {
	case req.ChannelID == 0:
		return replyBadRequest(c, "channel_id is required")
	case req.Title == "" || req.Content == "":
		return replyBadRequest(c, "Title and content are required")
	case validation.Length(req.Title) > validation.MaxPostTitle:
		return replyBadRequest(c, "Title must be at most 200 characters long")
	case validation.Length(req.Content) > validation.MaxPostContent:
		return replyBadRequest(c, "Content must be at most 65536 characters long")
	}

	post, err := h.postService.Create(c.Context(), userID, req.ChannelID, req.Title, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrChannelNotFound) {
			return replyErr(c, fiber.StatusNotFound, "Channel does not exist")
		}
		return replyInternal(c, err, "Failed to create post")
	}

	return c.Status(fiber.StatusCreated).JSON(post)
}

func (h *PostHandler) List(c fiber.Ctx) error {
	filter := service.PostFilter{}

	if raw := c.Query("channel_id"); raw != "" {
		channelID, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return replyBadRequest(c, "channel_id must be a number")
		}
		id := uint(channelID)
		filter.ChannelID = &id
	}

	query, err := searchQuery(c)
	if err != nil {
		return replyBadRequest(c, err.Error())
	}
	filter.Query = query

	limit, offset, err := parsePagination(c)
	if err != nil {
		return replyBadRequest(c, err.Error())
	}
	filter.Limit = limit
	filter.Offset = offset

	posts, err := h.postService.List(c.Context(), filter)
	if err != nil {
		return replyInternal(c, err, "Failed to fetch posts")
	}

	// Total is counted without the paging clause so the client can render
	// a page count even when this page is the last one.
	total, err := h.postService.Count(c.Context(), filter)
	if err != nil {
		return replyInternal(c, err, "Failed to count posts")
	}

	return c.JSON(fiber.Map{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"items":  posts,
	})
}

func (h *PostHandler) GetByID(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return replyBadRequest(c, "Post id must be a number")
	}

	post, err := h.postService.GetByID(c.Context(), uint(id))
	if err != nil {
		return svcErr(c, err, "Failed to fetch post")
	}

	return c.JSON(post)
}

func (h *PostHandler) Update(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return replyBadRequest(c, "Post id must be a number")
	}

	var req UpdatePostRequest

	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	if req.Title == nil && req.Content == nil {
		return replyBadRequest(c, "Nothing to update: provide title and/or content")
	}

	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			return replyBadRequest(c, "Title cannot be empty")
		}
		if validation.Length(trimmed) > validation.MaxPostTitle {
			return replyBadRequest(c, "Title must be at most 200 characters long")
		}
		req.Title = &trimmed
	}

	if req.Content != nil {
		trimmed := strings.TrimSpace(*req.Content)
		if trimmed == "" {
			return replyBadRequest(c, "Content cannot be empty")
		}
		if validation.Length(trimmed) > validation.MaxPostContent {
			return replyBadRequest(c, "Content must be at most 65536 characters long")
		}
		req.Content = &trimmed
	}

	post, err := h.postService.Update(c.Context(), uint(id), userID, currentIsAdmin(c), service.UpdatePostInput{
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
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return replyBadRequest(c, "Post id must be a number")
	}

	if err := h.postService.Delete(c.Context(), uint(id), userID, currentIsAdmin(c)); err != nil {
		return postErrorResponse(c, err, "Failed to delete post")
	}

	return c.JSON(fiber.Map{"deleted": true})
}

// postErrorResponse maps the service errors shared by update/delete to HTTP
// responses. It is svcErr plus the "you can only touch your own" case.
func postErrorResponse(c fiber.Ctx, err error, internalMessage string) error {
	if errors.Is(err, service.ErrForbidden) {
		return replyErr(c, fiber.StatusForbidden, "You can only modify your own posts")
	}
	return svcErr(c, err, internalMessage)
}
