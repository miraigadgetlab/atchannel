package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

const (
	defaultPageLimit = 20
	maxPageLimit     = 100
)

// ErrBadLimit and ErrBadOffset are returned by parsePagination for bad input.
var (
	ErrBadLimit  = errors.New("limit must be a number between 1 and 100")
	ErrBadOffset = errors.New("offset must be a non-negative number")
)

// parsePagination reads the limit/offset query params, applying defaults.
func parsePagination(c fiber.Ctx) (limit, offset int, err error) {
	limit = defaultPageLimit

	if raw := c.Query("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > maxPageLimit {
			return 0, 0, ErrBadLimit
		}
	}

	if raw := c.Query("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 {
			return 0, 0, ErrBadOffset
		}
	}

	return limit, offset, nil
}
