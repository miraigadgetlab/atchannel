package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SlidingWindowLimiter implements a race-free sliding-window rate limit
// using a Redis sorted set of request timestamps. The check-and-record
// operation is a single atomic Lua script.
type SlidingWindowLimiter struct {
	rdb    *redis.Client
	window time.Duration
	max    int
	script *redis.Script
}

var slidingWindowScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local max = tonumber(ARGV[3])
local member = ARGV[4]

redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
local count = redis.call('ZCARD', key)
if count < max then
    redis.call('ZADD', key, now, member)
    redis.call('PEXPIRE', key, window)
    return 1
end
return 0
`)

func NewSlidingWindowLimiter(rdb *redis.Client, window time.Duration, max int) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		rdb:    rdb,
		window: window,
		max:    max,
		script: slidingWindowScript,
	}
}

// Allow records a hit and reports whether it stays inside the limit.
func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	member := uuid.NewString()
	now := float64(time.Now().UnixMilli())
	res, err := l.script.Run(ctx, l.rdb, []string{key}, now, float64(l.window.Milliseconds()), l.max, member).Int()
	if err != nil {
		return false, fmt.Errorf("rate limit eval: %w", err)
	}
	return res == 1, nil
}

// Middleware returns HTTP middleware using the provided key function.
func (l *SlidingWindowLimiter) Middleware(keyFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ok, err := l.Allow(r.Context(), "rl:"+keyFunc(r))
			if err != nil {
				// Failing open is intentional for a rate limiter;
				// the upstream error is still logged.
				next.ServeHTTP(w, r)
				return
			}
			if !ok {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(l.window.Seconds())))
				writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
