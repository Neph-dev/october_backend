package http

import (
	"net/http"
	"time"

	"github.com/Neph-dev/october_backend/internal/interfaces/http/middleware"
	"github.com/Neph-dev/october_backend/pkg/logger"
)

// Router handles HTTP routing for the application
type Router struct {
	logger logger.Logger
	mux    *http.ServeMux
}

func NewRouter(logger logger.Logger) *Router {
	return &Router{
		logger: logger,
		mux:    http.NewServeMux(),
	}
}

// SetupRoutes configures all application routes
func (r *Router) SetupRoutes() {
	// Health check endpoint
	r.mux.HandleFunc("/health", r.handleHealth)
	
	// API routes would go here
	// r.mux.HandleFunc("/api/v1/news", r.handleNews)
}

// ServeHTTP implements http.Handler interface with middleware chain
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Create middleware chain
	handler := middleware.Recovery(r.logger)(
		middleware.RequestLogger(r.logger)(r.mux),
	)
	
	handler.ServeHTTP(w, req)
}

// handleHealth handles health check requests
func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}
