package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// clientKey resolves the rate limit bucket for a request.
//
// Behind a reverse proxy every request arrives from the proxy's address, so
// c.IP() alone would put all users in one bucket. Fiber resolves this: when
// Config.TrustProxy is set (see TRUSTED_PROXIES) c.IP() walks the
// X-Forwarded-For chain and stops at the first untrusted hop, so a client
// cannot forge its way out of its own bucket.
func clientKey(c fiber.Ctx) string {
	return c.IP()
}

// NewAuthRateLimiter throttles a public auth endpoint per client IP.
//
// skipSuccess makes successful attempts refund their slot, so only failed
// ones keep counting toward the budget: a user logging in all day never
// trips the limit, a password guessing run does. Blocked callers get a
// 429 with Retry-After (set by the limiter itself).
func NewAuthRateLimiter(max int, window time.Duration, skipSuccess bool) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:                    max,
		Expiration:             window,
		SkipSuccessfulRequests: skipSuccess,
		KeyGenerator:           clientKey,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Too Many Requests",
				"message": "Too many attempts from this address, try again later",
			})
		},
	})
}

// NewWriteRateLimiter throttles content creation (posts, comments, channels)
// per authenticated user. Every attempt counts — there is no "success" to
// refund — because the goal is capping how fast one account can flood the
// site, not catching mistakes.
func NewWriteRateLimiter(max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		KeyGenerator: func(c fiber.Ctx) string {
			// Prefer the authenticated identity: one person on five devices
			// (or five IPs) should still share a single budget.
			if claims, ok := c.Locals("user").(*UserClaims); ok && claims != nil {
				return "u:" + claims.UserID
			}
			return "ip:" + clientKey(c)
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Too Many Requests",
				"message": "You are posting too quickly, slow down",
			})
		},
	})
}

// accountHits tracks failed logins for one account inside a window.
type accountHits struct {
	count int
	start time.Time
}

// AccountLimiter locks a single account after too many failed logins.
// Per-IP limits are bypassed by a botnet; this one follows the account,
// so credential stuffing against one target stays expensive. Counters
// live in memory: a restart clears them, which is acceptable for a
// brute-force brake (stateless deploys should move this to Redis).
type AccountLimiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	hits   map[string]*accountHits
}

func NewAccountLimiter(max int, window time.Duration) *AccountLimiter {
	return &AccountLimiter{
		max:    max,
		window: window,
		hits:   make(map[string]*accountHits),
	}
}

// Blocked reports how long the account must wait before trying again.
// The second return value is false when attempts are still allowed.
func (a *AccountLimiter) Blocked(email string) (time.Duration, bool) {
	key := normalizeEmail(email)

	a.mu.Lock()
	defer a.mu.Unlock()

	h, ok := a.hits[key]
	if !ok {
		return 0, false
	}

	if time.Since(h.start) >= a.window {
		delete(a.hits, key)
		return 0, false
	}

	if h.count < a.max {
		return 0, false
	}

	return a.window - time.Since(h.start), true
}

// Fail records a rejected login attempt.
func (a *AccountLimiter) Fail(email string) {
	key := normalizeEmail(email)

	a.mu.Lock()
	defer a.mu.Unlock()

	// Opportunistic sweep so the map cannot grow without bound even if
	// attackers spray thousands of distinct addresses.
	if len(a.hits) > 10000 {
		for k, h := range a.hits {
			if time.Since(h.start) >= a.window {
				delete(a.hits, k)
			}
		}
	}

	h, ok := a.hits[key]
	if !ok || time.Since(h.start) >= a.window {
		a.hits[key] = &accountHits{count: 1, start: time.Now()}
		return
	}

	h.count++
}

// Clear forgets the counter after a successful login, so legitimate
// users who typo a password once do not carry the penalty forever.
func (a *AccountLimiter) Clear(email string) {
	key := normalizeEmail(email)

	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.hits, key)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
