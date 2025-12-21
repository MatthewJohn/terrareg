package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter represents a rate limiter that can be used to limit request rates
type RateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
	mu       sync.Mutex
}

// RateLimiterMiddleware manages rate limiting for multiple clients
type RateLimiterMiddleware struct {
	clients map[string]*RateLimiter
	mu      sync.RWMutex
	rate    rate.Limit
	burst   int
}

// NewRateLimiterMiddleware creates a new rate limiter middleware
func NewRateLimiterMiddleware(requestsPerSecond float64, burstSize int) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		clients: make(map[string]*RateLimiter),
		rate:    rate.Limit(requestsPerSecond),
		burst:   burstSize,
	}
}

// getClientLimiter gets or creates a rate limiter for a specific client
func (r *RateLimiterMiddleware) getClientLimiter(clientIP string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	if limiter, exists := r.clients[clientIP]; exists {
		limiter.mu.Lock()
		defer limiter.mu.Unlock()

		limiter.lastSeen = time.Now()
		return limiter.limiter
	}

	limiter := &RateLimiter{
		limiter:  rate.NewLimiter(r.rate, r.burst),
		lastSeen: time.Now(),
	}

	r.clients[clientIP] = limiter

	// Start cleanup goroutine
	go r.cleanupClients()

	return limiter.limiter
}

// cleanupClients removes old rate limiters to prevent memory leaks
func (r *RateLimiterMiddleware) cleanupClients() {
	for {
		time.Sleep(time.Minute)

		r.mu.Lock()
		for ip, limiter := range r.clients {
			limiter.mu.Lock()

			if time.Since(limiter.lastSeen) > 5*time.Minute {
				delete(r.clients, ip)
			}

			limiter.mu.Unlock()
		}
		r.mu.Unlock()
	}
}

// RateLimitAuth returns a middleware that rate limits authentication endpoints
func (r *RateLimiterMiddleware) RateLimitAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			clientIP := getClientIP(req)

			limiter := r.getClientLimiter(clientIP)

			if !limiter.Allow() {
				http.Error(w, "Too many authentication attempts", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(req *http.Request) string {
	// Check X-Forwarded-For header first (for load balancers)
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := len(xff); idx > 0 {
			if commaIdx := 0; commaIdx < len(xff) && xff[commaIdx] != ','; commaIdx++ {
				// Find the first comma or end of string
				for commaIdx < len(xff) && xff[commaIdx] != ',' && xff[commaIdx] != ' ' {
					commaIdx++
				}
				return xff[:commaIdx]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return req.RemoteAddr
}

// RateLimitAuthSimple is a simple rate limiter that doesn't track individual clients
func RateLimitAuthSimple(requestsPerSecond float64, burst int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}