package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Rate limiting implementation
type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests
	per      time.Duration // time window
	ctx      context.Context
	cancel   context.CancelFunc
}

type visitor struct {
	lastSeen time.Time
	count    int
}

func newRateLimiter(rate int, per time.Duration) *rateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		per:      per,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Clean up old visitors every minute
	go rl.cleanupVisitors()

	return rl
}

func (rl *rateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > rl.per {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.ctx.Done():
			// Graceful shutdown: cleanup goroutine exits when context is cancelled
			log.Println("🛑 Rate limiter cleanup goroutine shutting down")
			return
		}
	}
}

// Shutdown gracefully stops the rate limiter cleanup goroutine
func (rl *rateLimiter) Shutdown() {
	rl.cancel()
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	now := time.Now()

	if !exists {
		rl.visitors[ip] = &visitor{lastSeen: now, count: 1}
		return true
	}

	// Reset count if time window has passed
	if now.Sub(v.lastSeen) > rl.per {
		v.count = 1
		v.lastSeen = now
		return true
	}

	// Check if rate limit exceeded
	if v.count >= rl.rate {
		return false
	}

	v.count++
	v.lastSeen = now
	return true
}

// RateLimitMiddleware creates a rate limiting middleware and returns the limiter for shutdown
func RateLimitMiddleware(rate int, per time.Duration) (*rateLimiter, gin.HandlerFunc) {
	limiter := newRateLimiter(rate, per)

	handler := func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.allow(ip) {
			log.Printf("⚠️  Rate limit exceeded for IP: %s", ip)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}

	return limiter, handler
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// RequestSizeLimitMiddleware limits the size of request bodies
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Strict Transport Security (if using HTTPS)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// LoggerMiddleware provides request logging with request ID
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		requestID, _ := c.Get("request_id")

		log.Printf("[%s] %s %s - Status: %d - Duration: %v - IP: %s",
			requestID, method, path, statusCode, duration, c.ClientIP())
	}
}

// ErrorHandlerMiddleware handles errors and prevents information leakage
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there were any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			requestID, _ := c.Get("request_id")

			// Log the actual error with request ID
			log.Printf("[%s] Error: %v", requestID, err.Error())

			// Don't expose internal errors to client in production
			if !isTestMode() && !isDevMode() {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "An error occurred while processing your request",
					"request_id": requestID,
				})
			}
		}
	}
}

// isDevMode checks if running in development mode
func isDevMode() bool {
	return gin.Mode() == gin.DebugMode
}
