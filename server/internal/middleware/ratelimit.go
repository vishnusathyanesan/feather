package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := rl.getKey(r)
		ctx := r.Context()

		count, err := rl.increment(ctx, key)
		if err != nil {
			// If Redis is down, allow the request
			next.ServeHTTP(w, r)
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
	if userID := GetUserID(r.Context()); userID.String() != "00000000-0000-0000-0000-000000000000" {
		return fmt.Sprintf("ratelimit:%s", userID.String())
	}
	return fmt.Sprintf("ratelimit:%s", r.RemoteAddr)
}

func (rl *RateLimiter) increment(ctx context.Context, key string) (int64, error) {
	pipe := rl.redis.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, rl.window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}
