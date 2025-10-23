package http

import (
	"net/http"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/internal/interfaces/http/handlers"
	"github.com/Neph-dev/october_backend/internal/interfaces/http/middleware"
	"github.com/Neph-dev/october_backend/pkg/logger"
)

// Router handles HTTP routing for the application
type Router struct {
	logger         logger.Logger
	mux            *http.ServeMux
	companyHandler *handlers.CompanyHandler
	rateLimiter    *middleware.RateLimiter
}

func NewRouter(logger logger.Logger, companyService company.Service) *Router {
	// Create rate limiter: 10 requests per second, burst of 20
	rateLimiter := middleware.NewRateLimiter(10.0, 20, logger)
	
	return &Router{
		logger:         logger,
		mux:            http.NewServeMux(),
		companyHandler: handlers.NewCompanyHandler(companyService, logger),
		rateLimiter:    rateLimiter,
	}
}

// SetupRoutes configures all application routes
func (r *Router) SetupRoutes() {
	r.mux.HandleFunc("/health", r.handleHealth)
	
	// Company API routes with rate limiting
	r.mux.HandleFunc("/company/", r.handleCompanyByName)
	r.mux.HandleFunc("/companies", r.handleCompanies)
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

// GET /company/{company-name} with rate limiting
func (r *Router) handleCompanyByName(w http.ResponseWriter, req *http.Request) {
	// Only allow GET requests
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.companyHandler.GetCompanyByName))
	rateLimitedHandler.ServeHTTP(w, req)
}

// handleCompanies handles company collection operations (for seeding data)
func (r *Router) handleCompanies(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		// Create company (for seeding data)
		r.companyHandler.CreateCompany(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
