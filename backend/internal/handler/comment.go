package handler

import (
	"errors"
	"strconv"
	"strings"

	"atchannel-backend/internal/service"
	"atchannel-backend/internal/validation"

	"github.com/gofiber/fiber/v3"
)

const maxCommentLength = validation.MaxCommentContent

type CreateCommentRequest struct {
	Content string `json:"content"`
}

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(cs *service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: cs}
}

func (h *CommentHandler) Create(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	postID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return replyBadRequest(c, "Post id must be a number")
	}

	var req CreateCommentRequest

	if err := c.Bind().Body(&req); err != nil {
		return replyBadRequest(c, "Invalid request payload format")
	}

	req.Content = strings.TrimSpace(req.Content)

	switch {
	case req.Content == "":
		return replyBadRequest(c, "Content is required")
	case validation.Length(req.Content) > maxCommentLength:
		return replyBadRequest(c, "Comment must be at most 5000 characters long")
	}

	comment, err := h.commentService.Create(c.Context(), userID, uint(postID), req.Content)
	if err != nil {
		return svcErr(c, err, "Failed to create comment")
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

func (h *CommentHandler) List(c fiber.Ctx) error {
	postID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return replyBadRequest(c, "Post id must be a number")
	}

	limit, offset, err := parsePagination(c)
	if err != nil {
		return replyBadRequest(c, err.Error())
	}

	comments, total, err := h.commentService.ListByPost(c.Context(), uint(postID), limit, offset)
	if err != nil {
		return svcErr(c, err, "Failed to fetch comments")
	}

	return c.JSON(fiber.Map{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"items":  comments,
	})
}

func (h *CommentHandler) Delete(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	commentID, err := strconv.ParseUint(c.Params("commentID"), 10, 64)
	if err != nil {
		return replyBadRequest(c, "Comment id must be a number")
	}

	if err := h.commentService.Delete(c.Context(), uint(commentID), userID, currentIsAdmin(c)); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			return replyErr(c, fiber.StatusForbidden, "You can only delete your own comments")
		}
		return svcErr(c, err, "Failed to delete comment")
	}

	return c.JSON(fiber.Map{"deleted": true})
}
