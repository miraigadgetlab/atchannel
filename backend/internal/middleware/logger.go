package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
)

// RequestLogger prints one line per request: who asked, what, how long it
// took, and the request id (when RequestID middleware ran before it). The
// client IP is included because it is the only identity available on
// unauthenticated routes, and 4xx/5xx lines are the ones worth grepping.
// Every line carries id=<request-id> so a client-reported error can be
// matched to its log entry even when thousands arrive per second.
func RequestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		status := c.Response().StatusCode()
		id, _ := c.Locals(RequestIDKey).(string)

		switch {
		case status >= 500:
			log.Printf("ERROR  %s %s %d %s ip=%s id=%s err=%v",
				c.Method(), c.Path(), status, time.Since(start).Round(time.Microsecond), c.IP(), id, err)
		case status >= 400:
			log.Printf("WARN   %s %s %d %s ip=%s id=%s",
				c.Method(), c.Path(), status, time.Since(start).Round(time.Microsecond), c.IP(), id)
		default:
			log.Printf("%-6s %s %d %s ip=%s id=%s",
				c.Method(), c.Path(), status, time.Since(start).Round(time.Microsecond), c.IP(), id)
		}

		return err
	}
}
