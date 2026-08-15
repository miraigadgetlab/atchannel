package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/kosero/atchannel/backend/internal/middleware"
)

func newTestLimiter(t *testing.T, window time.Duration, max int) (*middleware.SlidingWindowLimiter, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return middleware.NewSlidingWindowLimiter(rdb, window, max), rdb
}

func TestSlidingWindowAllowsWithinLimit(t *testing.T) {
	limiter, _ := newTestLimiter(t, time.Minute, 3)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		ok, err := limiter.Allow(ctx, "test:key")
		if err != nil {
			t.Fatalf("allow %d: %v", i, err)
		}
		if !ok {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	// Fourth within the window is rejected.
	ok, err := limiter.Allow(ctx, "test:key")
	if err != nil {
		t.Fatalf("allow: %v", err)
	}
	if ok {
		t.Fatal("request beyond limit should be rejected")
	}
}

func TestSlidingWindowSeparateKeys(t *testing.T) {
	limiter, _ := newTestLimiter(t, time.Minute, 1)
	ctx := context.Background()

	ok, _ := limiter.Allow(ctx, "a")
	if !ok {
		t.Fatal("first allow for a")
	}
	ok, _ = limiter.Allow(ctx, "b")
	if !ok {
		t.Fatal("b should have its own window")
	}
	ok, _ = limiter.Allow(ctx, "a")
	if ok {
		t.Fatal("a already exhausted its window")
	}
}

func TestSlidingWindowSlidesAfterWindow(t *testing.T) {
	limiter, _ := newTestLimiter(t, 10*time.Millisecond, 1)
	ctx := context.Background()

	ok, _ := limiter.Allow(ctx, "k")
	if !ok {
		t.Fatal("first allow")
	}
	time.Sleep(30 * time.Millisecond)
	ok, _ = limiter.Allow(ctx, "k")
	if !ok {
		t.Fatal("window should have slid; request allowed")
	}
}

func TestSlidingWindowMiddlewareRejects(t *testing.T) {
	limiter, _ := newTestLimiter(t, time.Minute, 1)
	mw := limiter.Middleware(func(r *http.Request) string { return "m" })
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/", nil))
	if first.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", first.Code)
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", second.Code)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}
