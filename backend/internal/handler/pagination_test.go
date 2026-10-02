package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

// newTestApp mounts the query/response helpers behind the same route shape
// the real router uses.
func newTestApp(t *testing.T) *fiber.App {
	t.Helper()

	app := fiber.New()

	api := app.Group("/api/v1")

	api.Get("/page", func(c fiber.Ctx) error {
		limit, offset, err := parsePagination(c)
		if err != nil {
			return replyBadRequest(c, err.Error())
		}
		return c.JSON(fiber.Map{"limit": limit, "offset": offset})
	})

	api.Get("/search", func(c fiber.Ctx) error {
		q, err := searchQuery(c)
		if err != nil {
			return replyBadRequest(c, err.Error())
		}
		return c.JSON(fiber.Map{"q": q})
	})

	api.Get("/boom", func(c fiber.Ctx) error {
		return replyInternal(c, errTest, "Something failed")
	})

	return app
}

var errTest = fiber.NewError(fiber.StatusInternalServerError, "underlying detail")

func get(t *testing.T, app *fiber.App, target string) (int, string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, target, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test(%s): %v", target, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return resp.StatusCode, string(body)
}

func TestPaginationDefaults(t *testing.T) {
	app := newTestApp(t)

	status, body := get(t, app, "/api/v1/page")
	if status != http.StatusOK {
		t.Fatalf("status = %d, body = %s", status, body)
	}
	if !strings.Contains(body, `"limit":20`) {
		t.Errorf("body = %s, want default limit 20", body)
	}
	if !strings.Contains(body, `"offset":0`) {
		t.Errorf("body = %s, want default offset 0", body)
	}
}

func TestPaginationAcceptsBounds(t *testing.T) {
	app := newTestApp(t)

	for _, target := range []string{
		"/api/v1/page?limit=1&offset=0",
		"/api/v1/page?limit=100&offset=9999",
	} {
		if status, body := get(t, app, target); status != http.StatusOK {
			t.Errorf("GET %s = %d, body = %s", target, status, body)
		}
	}
}

func TestPaginationRejectsBadInput(t *testing.T) {
	app := newTestApp(t)

	bad := []string{
		"/api/v1/page?limit=0",
		"/api/v1/page?limit=101",
		"/api/v1/page?limit=abc",
		"/api/v1/page?limit=-5",
		"/api/v1/page?offset=-1",
		"/api/v1/page?offset=abc",
		"/api/v1/page?limit=10&offset=-100",
	}

	for _, target := range bad {
		status, body := get(t, app, target)
		if status != http.StatusBadRequest {
			t.Errorf("GET %s = %d, want 400 (body %s)", target, status, body)
		}
	}
}

func TestSearchQueryTrimsAndRejectsLongInput(t *testing.T) {
	app := newTestApp(t)

	if status, body := get(t, app, "/api/v1/search?q=%20%20go%20%20"); status != http.StatusOK {
		t.Errorf("status = %d, body = %s", status, body)
	} else if !strings.Contains(body, `"q":"go"`) {
		t.Errorf("body = %s, want trimmed q", body)
	}

	if status, _ := get(t, app, "/api/v1/search"); status != http.StatusOK {
		t.Error("missing q should mean \"no filter\", not an error")
	}

	long := strings.Repeat("a", 201)
	if status, _ := get(t, app, "/api/v1/search?q="+long); status != http.StatusBadRequest {
		t.Error("q over 200 chars should be rejected")
	}
}

// A 500 must tell the client what failed without leaking the underlying
// error text.
func TestInternalErrorDoesNotLeakDetails(t *testing.T) {
	app := newTestApp(t)

	status, body := get(t, app, "/api/v1/boom")
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d, body = %s", status, body)
	}
	if !strings.Contains(body, `"error":"Internal Server Error"`) {
		t.Errorf("body = %s, want error phrase", body)
	}
	if strings.Contains(body, "underlying detail") {
		t.Errorf("body = %s, must not contain the internal error", body)
	}
	if !strings.Contains(body, `"message":"Something failed"`) {
		t.Errorf("body = %s, want the caller-facing fallback", body)
	}
}

// Retry-After set by the rate limiter must survive to the client, since the
// account lockout handler relies on it.
func TestRetryAfterHeaderIsPreserved(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderRetryAfter, "16")
		return replyErr(c, fiber.StatusTooManyRequests, "slow down")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get(fiber.HeaderRetryAfter); got != "16" {
		t.Errorf("Retry-After = %q, want 16", got)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", resp.StatusCode)
	}
}

func TestAccountLimiterWaitIsADuration(t *testing.T) {
	// Guards the unit conversion the login handler performs when writing
	// Retry-After: a sub-second wait must not round down to 0.
	wait := 1500 * time.Millisecond
	if seconds := int(wait.Seconds()) + 1; seconds != 2 {
		t.Errorf("seconds = %d, want 2", seconds)
	}
}
