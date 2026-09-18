package middleware

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const tokenBucketScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local ttl_ms = tonumber(ARGV[5])

local bucket = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(bucket[1])
local ts = tonumber(bucket[2])

if tokens == nil then
    tokens = capacity
    ts = now
end

local elapsed = now - ts
if elapsed < 0 then
    elapsed = 0
end

tokens = math.min(capacity, tokens + (elapsed * rate))

local allowed = 0
if tokens >= requested then
    tokens = tokens - requested
    allowed = 1
end

redis.call("HMSET", key, "tokens", tokens, "ts", now)
redis.call("PEXPIRE", key, ttl_ms)

return allowed
`

const rateLimitCallTimeout = 750 * time.Millisecond

type IPRateLimiter struct {
	client   *redis.Client
	script   *redis.Script
	rate     float64
	burst    int
	ttl      time.Duration
	prefix   string
	failOpen bool
}

func NewIPRateLimiter(client *redis.Client, requestsPerMinute, burst int, ttl time.Duration, failOpen bool) *IPRateLimiter {
	return &IPRateLimiter{
		client:   client,
		script:   redis.NewScript(tokenBucketScript),
		rate:     float64(requestsPerMinute) / 60.0,
		burst:    burst,
		ttl:      ttl,
		prefix:   "ratelimit:",
		failOpen: failOpen,
	}
}

func (rl *IPRateLimiter) Allow(key string) bool {
	return rl.AllowCtx(context.Background(), key)
}

func (rl *IPRateLimiter) AllowCtx(ctx context.Context, key string) bool {
	ctx, cancel := context.WithTimeout(ctx, rateLimitCallTimeout)
	defer cancel()

	now := float64(time.Now().UnixNano()) / 1e9
	ttlMs := rl.ttl.Milliseconds()
	if ttlMs <= 0 {
		ttlMs = 60000
	}

	res, err := rl.script.Run(ctx, rl.client,
		[]string{rl.prefix + key},
		rl.rate, rl.burst, now, 1, ttlMs,
	).Result()

	if err != nil {
		if rl.failOpen {
			log.Printf("ratelimit: redis error, failing open: %v", err)
			return true
		}
		log.Printf("ratelimit: redis error, failing closed: %v", err)
		return false
	}

	allowed, ok := res.(int64)
	if !ok {
		if rl.failOpen {
			log.Printf("ratelimit: unexpected script result type, failing open")
			return true
		}
		log.Printf("ratelimit: unexpected script result type, failing closed")
		return false
	}

	return allowed == 1
}

func (rl *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP() + "|" + c.FullPath()

		if !rl.AllowCtx(c.Request.Context(), key) {
			c.Header("Retry-After", "60")
			RateLimited(c, "")
			c.Abort()
			return
		}
		c.Next()
	}
}

func PerRoute(client *redis.Client, requestsPerMinute, burst int, ttl time.Duration, failOpen bool) gin.HandlerFunc {
	return NewIPRateLimiter(client, requestsPerMinute, burst, ttl, failOpen).Middleware()
}
