package http

import (
	"net/http"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/ai"
	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/internal/domain/news"
	"github.com/Neph-dev/october_backend/internal/interfaces/http/handlers"
	"github.com/Neph-dev/october_backend/internal/interfaces/http/middleware"
	"github.com/Neph-dev/october_backend/pkg/logger"
	"github.com/gorilla/mux"
)

// Router handles HTTP routing for the application
type Router struct {
	logger         logger.Logger
	router         *mux.Router
	companyHandler *handlers.CompanyHandler
	newsHandler    *handlers.NewsHandler
	aiHandler      *handlers.AIHandler
	rateLimiter    *middleware.RateLimiter
}

func NewRouter(logger logger.Logger, companyService company.Service, newsService *news.Service, aiService ai.Service) *Router {
	// Create rate limiter: 10 requests per second, burst of 20
	rateLimiter := middleware.NewRateLimiter(10.0, 20, logger)
	
	return &Router{
		logger:         logger,
		router:         mux.NewRouter(),
		companyHandler: handlers.NewCompanyHandler(companyService, logger),
		newsHandler:    handlers.NewNewsHandler(newsService, logger.Unwrap()),
		aiHandler:      handlers.NewAIHandler(aiService, logger.Unwrap()),
		rateLimiter:    rateLimiter,
	}
}

// SetupRoutes configures all application routes
func (r *Router) SetupRoutes() {
	// Health check
	r.router.HandleFunc("/health", r.handleHealth).Methods("GET")
	
	// Company API routes with rate limiting
	r.router.HandleFunc("/company/{name}", r.handleCompanyByName).Methods("GET")
	r.router.HandleFunc("/companies", r.handleCompanies).Methods("POST")
	
	// News API routes with rate limiting
	r.router.HandleFunc("/news", r.handleNews).Methods("GET")
	r.router.HandleFunc("/news/{id}", r.handleNewsById).Methods("GET")
	r.router.HandleFunc("/news/company/{name}", r.handleNewsByCompany).Methods("GET")
	
	// AI/RAG API routes with rate limiting
	r.router.HandleFunc("/ai/query", r.handleAIQuery).Methods("POST")
	r.router.HandleFunc("/ai/analyze", r.handleAIAnalyze).Methods("POST")
	r.router.HandleFunc("/ai/web-search", r.handleAIWebSearch).Methods("POST")
}

// ServeHTTP implements http.Handler interface with middleware chain
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Create middleware chain
	handler := middleware.Recovery(r.logger)(
		middleware.RequestLogger(r.logger)(r.router),
	)
	
	handler.ServeHTTP(w, req)
}

// handleHealth handles health check requests
func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}

// handleCompanyByName handles GET /company/{name} with rate limiting
func (r *Router) handleCompanyByName(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.companyHandler.GetCompanyByName))
	rateLimitedHandler.ServeHTTP(w, req)
}

// handleCompanies handles company collection operations (for seeding data)
func (r *Router) handleCompanies(w http.ResponseWriter, req *http.Request) {
	r.companyHandler.CreateCompany(w, req)
}

// handleNews handles GET /news with rate limiting
func (r *Router) handleNews(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.newsHandler.GetNews))
	rateLimitedHandler.ServeHTTP(w, req)
}

// handleNewsById handles GET /news/{id} with rate limiting
func (r *Router) handleNewsById(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.newsHandler.GetNewsById))
	rateLimitedHandler.ServeHTTP(w, req)
}

// handleNewsByCompany handles GET /news/company/{name} with rate limiting
func (r *Router) handleNewsByCompany(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.newsHandler.GetNewsByCompany))
	rateLimitedHandler.ServeHTTP(w, req)
}

// handleAIQuery handles POST /ai/query with rate limiting
func (r *Router) handleAIQuery(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting (stricter for AI endpoints due to cost)
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.aiHandler.QueryHandler))
	rateLimitedHandler.ServeHTTP(w, req)
}

// handleAIAnalyze handles POST /ai/analyze with rate limiting
func (r *Router) handleAIAnalyze(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.aiHandler.AnalyzeQueryHandler))
	rateLimitedHandler.ServeHTTP(w, req)
}

// handleAIWebSearch handles POST /ai/web-search with rate limiting
func (r *Router) handleAIWebSearch(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.aiHandler.WebSearchHandler))
	rateLimitedHandler.ServeHTTP(w, req)
}
