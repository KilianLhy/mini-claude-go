package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64
	capacity float64
}

func newRateLimiter(ratePerSecond, capacity float64) *rateLimiter {
	rl := &rateLimiter{
		buckets:  make(map[string]*bucket),
		rate:     ratePerSecond,
		capacity: capacity,
	}
	go rl.evictLoop()
	return rl
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &bucket{tokens: rl.capacity - 1, lastSeen: now}
		return true
	}

	b.tokens += now.Sub(b.lastSeen).Seconds() * rl.rate
	if b.tokens > rl.capacity {
		b.tokens = rl.capacity
	}
	b.lastSeen = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (rl *rateLimiter) evictLoop() {
	for {
		time.Sleep(10 * time.Minute)
		rl.mu.Lock()
		for key, b := range rl.buckets {
			if time.Since(b.lastSeen) > 10*time.Minute {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

func rateLimitMiddleware(rl *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, errorBody("too many requests, slow down"))
			return
		}
		c.Next()
	}
}
