package httpapi

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// bucket is a tiny token-bucket per key (usually IP).
type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// limiter is a per-key token bucket with periodic cleanup.
// Not distributed — fine for a single-instance deployment. If you ever run
// multiple instances behind a load balancer, move this to a shared store.
type limiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64 // tokens per second
	capacity float64 // burst
}

func newLimiter(rate, burst float64) *limiter {
	l := &limiter{
		buckets:  map[string]*bucket{},
		rate:     rate,
		capacity: burst,
	}
	go l.cleanupLoop()
	return l
}

func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.capacity, lastSeen: now}
		l.buckets[key] = b
	}
	// Refill by elapsed time.
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens = minFloat(l.capacity, b.tokens+elapsed*l.rate)
	b.lastSeen = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (l *limiter) cleanupLoop() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for range t.C {
		cutoff := time.Now().Add(-10 * time.Minute)
		l.mu.Lock()
		for k, b := range l.buckets {
			if b.lastSeen.Before(cutoff) {
				delete(l.buckets, k)
			}
		}
		l.mu.Unlock()
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// rateLimit middleware. `keyFn` maps the request to a bucket key.
// Default: client IP (as seen by Gin).
func rateLimit(l *limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !l.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, slow down",
			})
			return
		}
		c.Next()
	}
}
