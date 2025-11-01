package http

import (
	"net/http"

	"github.com/Neph-dev/october_backend/internal/domain/ai"
	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/internal/domain/market"
	"github.com/Neph-dev/october_backend/internal/domain/news"
	"github.com/Neph-dev/october_backend/internal/interfaces/http/handlers"
	"github.com/Neph-dev/october_backend/internal/interfaces/http/middleware"
	"github.com/Neph-dev/october_backend/pkg/logger"
	"github.com/Neph-dev/october_backend/pkg/metrics"
	"github.com/gorilla/mux"
)

// Router handles HTTP routing for the application
type Router struct {
	logger         logger.Logger
	router         *mux.Router
	companyHandler *handlers.CompanyHandler
	newsHandler    *handlers.NewsHandler
	aiHandler      *handlers.AIHandler
	marketHandler  *handlers.MarketHandler
	metricsHandler *handlers.MetricsHandler
	rateLimiter    *middleware.RateLimiter
	collector      *metrics.MetricsCollector
}

func NewRouter(logger logger.Logger, companyService company.Service, newsService *news.Service, aiService ai.Service, marketService market.Service, collector *metrics.MetricsCollector, version string) *Router {
	// Create rate limiter: 10 requests per second, burst of 20
	rateLimiter := middleware.NewRateLimiter(10.0, 20, logger)
	
	return &Router{
		logger:         logger,
		router:         mux.NewRouter(),
		companyHandler: handlers.NewCompanyHandler(companyService, logger),
		newsHandler:    handlers.NewNewsHandler(newsService, logger.Unwrap()),
		aiHandler:      handlers.NewAIHandler(aiService, logger.Unwrap()),
		marketHandler:  handlers.NewMarketHandler(marketService, logger.Unwrap()),
		metricsHandler: handlers.NewMetricsHandler(collector, newsService, aiService, companyService, logger, version),
		rateLimiter:    rateLimiter,
		collector:      collector,
	}
}

// SetupRoutes configures all application routes
func (r *Router) SetupRoutes() {
	// Monitoring and health endpoints (no rate limiting for these)
	r.router.HandleFunc("/health", r.metricsHandler.HandleHealth).Methods("GET")
	r.router.HandleFunc("/metrics", r.metricsHandler.HandleMetrics).Methods("GET")
	r.router.HandleFunc("/metrics.json", r.metricsHandler.HandleMetricsJSON).Methods("GET")
	r.router.HandleFunc("/readiness", r.metricsHandler.HandleReadiness).Methods("GET")
	r.router.HandleFunc("/liveness", r.metricsHandler.HandleLiveness).Methods("GET")
	
	// Company API routes with rate limiting
	r.router.HandleFunc("/companies", r.handleGetAllCompanies).Methods("GET")
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
	r.router.HandleFunc("/ai/summarise/{articleId}", r.handleAISummarizeArticle).Methods("GET")
	r.router.HandleFunc("/ai/cache/stats", r.handleAICacheStats).Methods("GET")
	
	// Market Data API routes with rate limiting
	r.router.HandleFunc("/market/quote/{ticker}", r.handleMarketQuote).Methods("GET")
	r.router.HandleFunc("/market/quotes", r.handleMarketQuotes).Methods("GET")
	r.router.HandleFunc("/market/tickers", r.handleMarketTickers).Methods("GET")
	r.router.HandleFunc("/market/status/{exchange}", r.handleMarketStatus).Methods("GET")
}

// ServeHTTP implements http.Handler interface with middleware chain
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Create middleware chain with metrics collection
	handler := middleware.Recovery(r.logger)(
		middleware.RequestLogger(r.logger)(
			middleware.MetricsMiddleware(r.collector)(r.router),
		),
	)
	
	handler.ServeHTTP(w, req)
}



// handleCompanyByName handles GET /company/{name} with rate limiting
func (r *Router) handleCompanyByName(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.companyHandler.GetCompanyByName))
	rateLimitedHandler.ServeHTTP(w, req)
}

// handleGetAllCompanies handles GET /companies with rate limiting
func (r *Router) handleGetAllCompanies(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.companyHandler.GetAllCompanies))
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

// handleAISummarizeArticle handles GET /ai/summarise/{articleId} with rate limiting
func (r *Router) handleAISummarizeArticle(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.aiHandler.SummarizeArticleHandler))
	rateLimitedHandler.ServeHTTP(w, req)
}

// handleAICacheStats handles GET /ai/cache/stats with rate limiting
func (r *Router) handleAICacheStats(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.aiHandler.CacheStatsHandler))
	rateLimitedHandler.ServeHTTP(w, req)
}

// Market Data Handlers

// handleMarketQuote handles GET /market/quote/{ticker} with rate limiting
func (r *Router) handleMarketQuote(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.marketHandler.GetQuoteHandler))
	rateLimitedHandler.ServeHTTP(w, req)
}

// handleMarketQuotes handles GET /market/quotes with rate limiting
func (r *Router) handleMarketQuotes(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.marketHandler.GetQuotesHandler))
	rateLimitedHandler.ServeHTTP(w, req)
}

// handleMarketTickers handles GET /market/tickers with rate limiting
func (r *Router) handleMarketTickers(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.marketHandler.GetAvailableTickersHandler))
	rateLimitedHandler.ServeHTTP(w, req)
}

// handleMarketStatus handles GET /market/status/{exchange} with rate limiting
func (r *Router) handleMarketStatus(w http.ResponseWriter, req *http.Request) {
	// Apply rate limiting
	rateLimitedHandler := r.rateLimiter.Middleware()(http.HandlerFunc(r.marketHandler.GetMarketStatusHandler))
	rateLimitedHandler.ServeHTTP(w, req)
}
