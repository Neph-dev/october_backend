package utils

import "net/http"

// GetClientIP extracts the client IP address from the request
func GetClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP if multiple are present
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	if r.RemoteAddr != "" {
		// Remove port if present
		for i := 0; i < len(r.RemoteAddr); i++ {
			if r.RemoteAddr[i] == ':' {
				return r.RemoteAddr[:i]
			}
		}
		return r.RemoteAddr
	}

	// Default fallback
	return "unknown"
}