package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis  *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(redisClient *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		limit:  limit,
		window: window,
	}
}

// Lua script for atomic increment with expiry.
// Returns the current count after increment.
// Sets expiry only on first increment (when count becomes 1).
var incrWithExpireScript = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
    redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return count
`)

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := rl.getKey(r)
		ctx := r.Context()

		count, err := rl.increment(ctx, key)
		if err != nil {
			// If Redis is down, reject the request (fail closed)
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, rl.limit-int(count))))

		if int(count) > rl.limit {
			w.Header().Set("Retry-After", strconv.Itoa(int(rl.window.Seconds())))
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) getKey(r *http.Request) string {
	// Use user ID if authenticated, otherwise IP
	if userID := GetUserID(r.Context()); userID != uuid.Nil {
		return fmt.Sprintf("ratelimit:%s", userID.String())
	}
	return fmt.Sprintf("ratelimit:%s", r.RemoteAddr)
}

func (rl *RateLimiter) increment(ctx context.Context, key string) (int64, error) {
	windowMs := rl.window.Milliseconds()
	result, err := incrWithExpireScript.Run(ctx, rl.redis, []string{key}, windowMs).Int64()
	if err != nil {
		return 0, err
	}
	return result, nil
}
