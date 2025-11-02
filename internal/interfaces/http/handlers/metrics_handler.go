package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/ai"
	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/internal/domain/news"
	"github.com/Neph-dev/october_backend/pkg/logger"
	"github.com/Neph-dev/october_backend/pkg/metrics"
)

// MetricsHandler handles metrics and health endpoints
type MetricsHandler struct {
	collector      *metrics.MetricsCollector
	newsService    *news.Service
	aiService      ai.Service
	companyService company.Service
	logger         logger.Logger
	version        string
}

func NewMetricsHandler(
	collector *metrics.MetricsCollector,
	newsService *news.Service,
	aiService ai.Service,
	companyService company.Service,
	logger logger.Logger,
	version string,
) *MetricsHandler {
	return &MetricsHandler{
		collector:      collector,
		newsService:    newsService,
		aiService:      aiService,
		companyService: companyService,
		logger:         logger,
		version:        version,
	}
}

// HandleMetrics serves Prometheus-formatted metrics
func (h *MetricsHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	
	metrics := h.collector.GetPrometheusMetrics()
	w.Write([]byte(metrics))
}

// HandleHealth serves health check with detailed status
func (h *MetricsHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	
	// Check database health by trying to count companies
	dbHealthy := true
	_, err := h.companyService.ListCompanies(ctx, 1, 0)
	if err != nil {
		h.logger.Warn("Database health check failed", "error", err)
		dbHealthy = false
	}
	
	// Check AI service health (this is a simple check)
	aiHealthy := true
	// We could add a simple ping method to AI service, but for now assume healthy
	
	status := h.collector.GetHealthStatus(ctx, h.version, dbHealthy, aiHealthy)
	
	w.Header().Set("Content-Type", "application/json")
	
	// Set appropriate HTTP status code
	if status.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	json.NewEncoder(w).Encode(status)
}

// HandleReadiness serves readiness probe for Kubernetes
func (h *MetricsHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	
	// Check if we can connect to database
	_, err := h.companyService.ListCompanies(ctx, 1, 0)
	if err != nil {
		h.logger.Warn("Readiness check failed - database not ready", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "not ready",
			"reason":  "database not available",
			"timestamp": time.Now(),
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now(),
	})
}

// HandleLiveness serves liveness probe for Kubernetes
func (h *MetricsHandler) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - if we can respond, we're alive
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}

// HandleMetricsJSON serves metrics in JSON format
func (h *MetricsHandler) HandleMetricsJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	metrics := h.collector.GetMetrics()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics":   metrics,
		"timestamp": time.Now(),
		"version":   h.version,
	})
}