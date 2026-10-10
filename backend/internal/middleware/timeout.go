package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Timeout bounds how long a single request may run. Every handler hands
// c.Context() to GORM, so when the deadline passes the in-flight query is
// cancelled and its pool connection is released instead of being held by a
// slow or stuck statement.
func Timeout(d time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), d)
		defer cancel()

		c.SetContext(ctx)
		return c.Next()
	}
}
