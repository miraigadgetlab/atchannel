package handler

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"atchannel-backend/internal/middleware"
	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

const maxCommentLength = 5000

type CreateCommentRequest struct {
	Content string `json:"content"`
}

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(cs *service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: cs}
}

func currentUserID(c fiber.Ctx) (uint, error) {
	claims, ok := c.Locals("user").(*middleware.UserClaims)
	if !ok || claims == nil {
		return 0, errors.New("missing authentication context")
	}

	id, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid user identity in token")
	}

	return uint(id), nil
}

// currentIsAdmin reports whether the caller's token carries the admin role.
func currentIsAdmin(c fiber.Ctx) bool {
	claims, ok := c.Locals("user").(*middleware.UserClaims)
	if !ok || claims == nil {
		return false
	}

	return slices.Contains(claims.Roles, "admin")
}

func (h *CommentHandler) Create(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
	}

	postID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Post id must be a number",
		})
	}

	var req CreateCommentRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request payload format",
		})
	}

	req.Content = strings.TrimSpace(req.Content)

	switch {
	case req.Content == "":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Content is required",
		})
	case len(req.Content) > maxCommentLength:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Comment must be at most 5000 characters long",
		})
	}

	comment, err := h.commentService.Create(c.Context(), userID, uint(postID), req.Content)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Post does not exist",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create comment",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

func (h *CommentHandler) List(c fiber.Ctx) error {
	postID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Post id must be a number",
		})
	}

	comments, err := h.commentService.ListByPost(c.Context(), uint(postID))
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Post does not exist",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to fetch comments",
		})
	}

	return c.JSON(comments)
}

func (h *CommentHandler) Delete(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": err.Error(),
		})
	}

	commentID, err := strconv.ParseUint(c.Params("commentID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Comment id must be a number",
		})
	}

	if err := h.commentService.Delete(c.Context(), uint(commentID), userID, currentIsAdmin(c)); err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Comment not found",
			})
		case errors.Is(err, service.ErrForbidden):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Forbidden",
				"message": "You can only delete your own comments",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to delete comment",
			})
		}
	}

	return c.JSON(fiber.Map{"deleted": true})
}
