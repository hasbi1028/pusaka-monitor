package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter — simple in-memory token bucket per IP
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter buat rate limiter baru
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	// Cleanup setiap menit
	go func() {
		for range time.Tick(time.Minute) {
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for ip, times := range rl.requests {
		valid := times[:0]
		for _, t := range times {
			if now.Sub(t) <= rl.window {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = valid
		}
	}
}

// Allow cek apakah IP boleh request
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	// Cleanup dulu yang sudah expired
	valid := rl.requests[ip][:0]
	for _, t := range rl.requests[ip] {
		if now.Sub(t) <= rl.window {
			valid = append(valid, t)
		}
	}
	rl.requests[ip] = valid

	if len(valid) >= rl.limit {
		return false
	}
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

// RateLimit middleware — limit request per IP
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	rl := NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak request. Coba lagi dalam beberapa menit.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
