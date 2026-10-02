package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
)

// RequestLogger prints one line per request: who asked, what, and how long it
// took. The client IP is included because it is the only identity available
// on unauthenticated routes, and 4xx/5xx lines are the ones worth grepping.
func RequestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		status := c.Response().StatusCode()

		switch {
		case status >= 500:
			log.Printf("ERROR  %s %s %d %s ip=%s err=%v",
				c.Method(), c.Path(), status, time.Since(start).Round(time.Microsecond), c.IP(), err)
		case status >= 400:
			log.Printf("WARN   %s %s %d %s ip=%s",
				c.Method(), c.Path(), status, time.Since(start).Round(time.Microsecond), c.IP())
		default:
			log.Printf("%-6s %s %d %s ip=%s",
				c.Method(), c.Path(), status, time.Since(start).Round(time.Microsecond), c.IP())
		}

		return err
	}
}
