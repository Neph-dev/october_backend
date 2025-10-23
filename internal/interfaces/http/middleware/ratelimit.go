package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/Neph-dev/october_backend/internal/interfaces/http/utils"
	"github.com/Neph-dev/october_backend/pkg/logger"
)

type RateLimiter struct {
	clients map[string]*clientLimiter
	mu      sync.RWMutex
	rate    rate.Limit
	burst   int
	logger  logger.Logger
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(requestsPerSecond float64, burst int, logger logger.Logger) *RateLimiter {
	// Validate input parameters
	if requestsPerSecond <= 0 {
		requestsPerSecond = 1 // Safe default
	}
	if burst <= 0 {
		burst = 5 // Safe default
	}

	rl := &RateLimiter{
		clients: make(map[string]*clientLimiter),
		rate:    rate.Limit(requestsPerSecond),
		burst:   burst,
		logger:  logger,
	}

	// Start cleanup goroutine to remove old clients
	go rl.cleanupClients()

	return rl
}

// Middleware returns the rate limiting middleware function
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := utils.GetClientIP(r)

			if !rl.isAllowed(clientIP) {
				rl.logger.Warn("Rate limit exceeded",
					"client_ip", clientIP,
					"path", r.URL.Path,
					"method", r.Method,
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "rate limit exceeded", "message": "too many requests"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isAllowed checks if a request from the given IP is allowed
func (rl *RateLimiter) isAllowed(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.clients[clientIP]
	if !exists {
		limiter = &clientLimiter{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.clients[clientIP] = limiter
	} else {
		limiter.lastSeen = time.Now()
	}

	return limiter.limiter.Allow()
}

// cleanupClients removes clients that haven't been seen for a while
func (rl *RateLimiter) cleanupClients() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-30 * time.Minute) // Remove clients not seen for 30 minutes

		for ip, client := range rl.clients {
			if client.lastSeen.Before(cutoff) {
				delete(rl.clients, ip)
			}
		}

		rl.logger.Debug("Cleaned up rate limiter clients", "remaining_clients", len(rl.clients))
		rl.mu.Unlock()
	}
}