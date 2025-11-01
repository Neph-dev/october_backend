package metrics

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/Neph-dev/october_backend/pkg/logger"
)

// MetricsCollector collects and stores application metrics
type MetricsCollector struct {
	mu                  sync.RWMutex
	httpRequestsTotal   map[string]int64        // endpoint -> count
	httpRequestDuration map[string][]float64    // endpoint -> durations in ms
	activeConnections   int64
	totalErrors         int64
	databaseConnections int64
	rssProcessedTotal   int64
	aiQueriesTotal      int64
	cacheHits           int64
	cacheMisses         int64
	memoryUsage         runtime.MemStats
	startTime           time.Time
	logger              logger.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger logger.Logger) *MetricsCollector {
	collector := &MetricsCollector{
		httpRequestsTotal:   make(map[string]int64),
		httpRequestDuration: make(map[string][]float64),
		startTime:           time.Now(),
		logger:              logger,
	}
	
	// Start periodic memory collection
	go collector.collectMemoryStats()
	
	return collector
}

// RecordHTTPRequest records an HTTP request with its duration
func (m *MetricsCollector) RecordHTTPRequest(endpoint, method string, duration time.Duration, statusCode int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := method + "_" + endpoint
	m.httpRequestsTotal[key]++
	
	if m.httpRequestDuration[key] == nil {
		m.httpRequestDuration[key] = make([]float64, 0)
	}
	m.httpRequestDuration[key] = append(m.httpRequestDuration[key], float64(duration.Nanoseconds())/1e6) // Convert to milliseconds
	
	// Keep only last 1000 measurements to avoid memory growth
	if len(m.httpRequestDuration[key]) > 1000 {
		m.httpRequestDuration[key] = m.httpRequestDuration[key][len(m.httpRequestDuration[key])-1000:]
	}
	
	if statusCode >= 400 {
		m.totalErrors++
	}
}

// RecordActiveConnection increments active connections
func (m *MetricsCollector) RecordActiveConnection() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeConnections++
}

// RecordConnectionClosed decrements active connections
func (m *MetricsCollector) RecordConnectionClosed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.activeConnections > 0 {
		m.activeConnections--
	}
}

// RecordDatabaseConnection sets current database connections
func (m *MetricsCollector) RecordDatabaseConnection(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.databaseConnections = count
}

// RecordRSSProcessed increments RSS articles processed
func (m *MetricsCollector) RecordRSSProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rssProcessedTotal++
}

// RecordAIQuery increments AI queries processed
func (m *MetricsCollector) RecordAIQuery() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.aiQueriesTotal++
}

// RecordCacheHit increments cache hits
func (m *MetricsCollector) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheHits++
}

// RecordCacheMiss increments cache misses
func (m *MetricsCollector) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheMisses++
}

// GetMetrics returns current metrics in Prometheus format
func (m *MetricsCollector) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	uptime := time.Since(m.startTime).Seconds()
	
	metrics := map[string]interface{}{
		"uptime_seconds":             uptime,
		"http_requests_total":        m.httpRequestsTotal,
		"http_request_duration_avg":  m.calculateAverageDurations(),
		"active_connections":         m.activeConnections,
		"total_errors":               m.totalErrors,
		"database_connections":       m.databaseConnections,
		"rss_processed_total":        m.rssProcessedTotal,
		"ai_queries_total":           m.aiQueriesTotal,
		"cache_hits_total":           m.cacheHits,
		"cache_misses_total":         m.cacheMisses,
		"cache_hit_ratio":           m.calculateCacheHitRatio(),
		"memory_usage_bytes":         m.memoryUsage.Alloc,
		"memory_sys_bytes":           m.memoryUsage.Sys,
		"gc_runs_total":              m.memoryUsage.NumGC,
		"goroutines_count":           runtime.NumGoroutine(),
	}
	
	return metrics
}

// GetPrometheusMetrics returns metrics in Prometheus text format
func (m *MetricsCollector) GetPrometheusMetrics() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	uptime := time.Since(m.startTime).Seconds()
	
	result := "# HELP october_uptime_seconds Time since application started\n"
	result += "# TYPE october_uptime_seconds counter\n"
	result += fmt.Sprintf("october_uptime_seconds %.2f\n\n", uptime)
	
	result += "# HELP october_http_requests_total Total number of HTTP requests\n"
	result += "# TYPE october_http_requests_total counter\n"
	for endpoint, count := range m.httpRequestsTotal {
		result += fmt.Sprintf("october_http_requests_total{endpoint=\"%s\"} %d\n", endpoint, count)
	}
	result += "\n"
	
	result += "# HELP october_http_request_duration_ms Average HTTP request duration in milliseconds\n"
	result += "# TYPE october_http_request_duration_ms gauge\n"
	avgDurations := m.calculateAverageDurations()
	for endpoint, avg := range avgDurations {
		result += fmt.Sprintf("october_http_request_duration_ms{endpoint=\"%s\"} %.2f\n", endpoint, avg)
	}
	result += "\n"
	
	result += "# HELP october_active_connections Current number of active connections\n"
	result += "# TYPE october_active_connections gauge\n"
	result += fmt.Sprintf("october_active_connections %d\n\n", m.activeConnections)
	
	result += "# HELP october_errors_total Total number of errors\n"
	result += "# TYPE october_errors_total counter\n"
	result += fmt.Sprintf("october_errors_total %d\n\n", m.totalErrors)
	
	result += "# HELP october_database_connections Current number of database connections\n"
	result += "# TYPE october_database_connections gauge\n"
	result += fmt.Sprintf("october_database_connections %d\n\n", m.databaseConnections)
	
	result += "# HELP october_rss_processed_total Total number of RSS articles processed\n"
	result += "# TYPE october_rss_processed_total counter\n"
	result += fmt.Sprintf("october_rss_processed_total %d\n\n", m.rssProcessedTotal)
	
	result += "# HELP october_ai_queries_total Total number of AI queries processed\n"
	result += "# TYPE october_ai_queries_total counter\n"
	result += fmt.Sprintf("october_ai_queries_total %d\n\n", m.aiQueriesTotal)
	
	result += "# HELP october_cache_hits_total Total number of cache hits\n"
	result += "# TYPE october_cache_hits_total counter\n"
	result += fmt.Sprintf("october_cache_hits_total %d\n\n", m.cacheHits)
	
	result += "# HELP october_cache_misses_total Total number of cache misses\n"
	result += "# TYPE october_cache_misses_total counter\n"
	result += fmt.Sprintf("october_cache_misses_total %d\n\n", m.cacheMisses)
	
	result += "# HELP october_cache_hit_ratio Cache hit ratio (0-1)\n"
	result += "# TYPE october_cache_hit_ratio gauge\n"
	result += fmt.Sprintf("october_cache_hit_ratio %.4f\n\n", m.calculateCacheHitRatio())
	
	result += "# HELP october_memory_usage_bytes Current memory usage in bytes\n"
	result += "# TYPE october_memory_usage_bytes gauge\n"
	result += fmt.Sprintf("october_memory_usage_bytes %d\n\n", m.memoryUsage.Alloc)
	
	result += "# HELP october_memory_sys_bytes Total memory obtained from system\n"
	result += "# TYPE october_memory_sys_bytes gauge\n"
	result += fmt.Sprintf("october_memory_sys_bytes %d\n\n", m.memoryUsage.Sys)
	
	result += "# HELP october_gc_runs_total Total number of garbage collection runs\n"
	result += "# TYPE october_gc_runs_total counter\n"
	result += fmt.Sprintf("october_gc_runs_total %d\n\n", m.memoryUsage.NumGC)
	
	result += "# HELP october_goroutines_count Current number of goroutines\n"
	result += "# TYPE october_goroutines_count gauge\n"
	result += fmt.Sprintf("october_goroutines_count %d\n\n", runtime.NumGoroutine())
	
	return result
}

// calculateAverageDurations calculates average durations for each endpoint
func (m *MetricsCollector) calculateAverageDurations() map[string]float64 {
	result := make(map[string]float64)
	
	for endpoint, durations := range m.httpRequestDuration {
		if len(durations) == 0 {
			continue
		}
		
		sum := 0.0
		for _, duration := range durations {
			sum += duration
		}
		result[endpoint] = sum / float64(len(durations))
	}
	
	return result
}

// calculateCacheHitRatio calculates cache hit ratio
func (m *MetricsCollector) calculateCacheHitRatio() float64 {
	total := m.cacheHits + m.cacheMisses
	if total == 0 {
		return 0.0
	}
	return float64(m.cacheHits) / float64(total)
}

// collectMemoryStats periodically collects memory statistics
func (m *MetricsCollector) collectMemoryStats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		m.mu.Lock()
		runtime.ReadMemStats(&m.memoryUsage)
		m.mu.Unlock()
	}
}

// HealthStatus represents the health status of the application
type HealthStatus struct {
	Status        string                 `json:"status"`
	Timestamp     time.Time              `json:"timestamp"`
	Uptime        string                 `json:"uptime"`
	Version       string                 `json:"version"`
	Checks        map[string]interface{} `json:"checks"`
	Metrics       map[string]interface{} `json:"metrics,omitempty"`
}

// GetHealthStatus returns detailed health status
func (m *MetricsCollector) GetHealthStatus(ctx context.Context, version string, dbHealthy, aiHealthy bool) *HealthStatus {
	uptime := time.Since(m.startTime)
	
	checks := map[string]interface{}{
		"database": map[string]interface{}{
			"status": func() string {
				if dbHealthy {
					return "healthy"
				}
				return "unhealthy"
			}(),
			"connections": m.databaseConnections,
		},
		"ai_service": map[string]interface{}{
			"status": func() string {
				if aiHealthy {
					return "healthy"
				}
				return "unhealthy"
			}(),
			"queries_processed": m.aiQueriesTotal,
		},
		"cache": map[string]interface{}{
			"status":    "healthy",
			"hit_ratio": m.calculateCacheHitRatio(),
			"hits":      m.cacheHits,
			"misses":    m.cacheMisses,
		},
		"memory": map[string]interface{}{
			"status":           "healthy",
			"usage_mb":         float64(m.memoryUsage.Alloc) / 1024 / 1024,
			"system_mb":        float64(m.memoryUsage.Sys) / 1024 / 1024,
			"gc_runs":          m.memoryUsage.NumGC,
			"goroutines":       runtime.NumGoroutine(),
		},
	}
	
	// Determine overall status
	status := "healthy"
	if !dbHealthy || !aiHealthy {
		status = "unhealthy"
	}
	
	return &HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
		Version:   version,
		Checks:    checks,
		Metrics:   m.GetMetrics(),
	}
}