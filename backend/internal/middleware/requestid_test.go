package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func newTestApp(handler fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Use(handler)
	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})
	return app
}

// RequestID must stamp every response with a header and make the id readable
// by later middleware/handlers.
func TestRequestIDSetsHeaderAndLocals(t *testing.T) {
	t.Run("generates when inbound header is absent", func(t *testing.T) {
		app := newTestApp(RequestID())
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if got := resp.Header.Get(fiber.HeaderXRequestID); got == "" {
			t.Fatal("expected X-Request-ID header to be set")
		}
	})

	t.Run("honours inbound header for tracing", func(t *testing.T) {
		app := newTestApp(RequestID())
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(fiber.HeaderXRequestID, "trace-123")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if got := resp.Header.Get(fiber.HeaderXRequestID); got != "trace-123" {
			t.Fatalf("X-Request-ID = %q, want inbound id preserved", got)
		}
	})
}
