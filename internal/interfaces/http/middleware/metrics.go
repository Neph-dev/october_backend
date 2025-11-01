package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/Neph-dev/october_backend/pkg/metrics"
	"github.com/gorilla/mux"
)

// MetricsMiddleware creates middleware that collects HTTP metrics
func MetricsMiddleware(collector *metrics.MetricsCollector) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Track active connections
			collector.RecordActiveConnection()
			defer collector.RecordConnectionClosed()
			
			// Create a response writer wrapper to capture status code
			wrappedWriter := &responseWriter{
				ResponseWriter: w,
				statusCode:     200, // Default to 200
			}
			
			// Process the request
			next.ServeHTTP(wrappedWriter, r)
			
			// Record metrics
			duration := time.Since(start)
			endpoint := extractEndpoint(r)
			method := r.Method
			statusCode := wrappedWriter.statusCode
			
			collector.RecordHTTPRequest(endpoint, method, duration, statusCode)
		})
	}
}

// extractEndpoint extracts a normalized endpoint path for metrics
func extractEndpoint(r *http.Request) string {
	path := r.URL.Path
	
	// Normalize common patterns
	if strings.HasPrefix(path, "/companies/") {
		return "/companies/{id}"
	}
	if strings.HasPrefix(path, "/news/") && strings.Contains(path, "/summary") {
		return "/news/{id}/summary"
	}
	if strings.HasPrefix(path, "/market/") {
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			switch parts[2] {
			case "quote":
				return "/market/quote/{ticker}"
			case "quotes":
				return "/market/quotes"
			case "tickers":
				return "/market/tickers"
			case "status":
				return "/market/status/{exchange}"
			}
		}
		return "/market/*"
	}
	
	// Return the original path for simple endpoints
	switch path {
	case "/health", "/metrics", "/companies", "/news", "/ai/query":
		return path
	default:
		// For unknown paths, return a generic pattern
		if strings.Count(path, "/") > 2 {
			return "/unknown/{...}"
		}
		return path
	}
}