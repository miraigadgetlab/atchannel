package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gofiber/fiber/v3"
)

// RequestIDKey is where the request id is stored for other middleware and the
// ErrorHandler to read.
const RequestIDKey = "request_id"

// RequestID tags every request with an id, echoing it back in the X-Request-ID
// response header so a client can quote it in a bug report. An inbound
// X-Request-ID is honoured (not replaced), which keeps tracing working when a
// frontend proxy already stamped the request.
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get(fiber.HeaderXRequestID)
		if id == "" {
			id = newRequestID()
		}
		c.Set(fiber.HeaderXRequestID, id)
		c.Locals(RequestIDKey, id)
		return c.Next()
	}
}

func newRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(buf)
}
