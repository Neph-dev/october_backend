package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/internal/interfaces/http/utils"
	"github.com/Neph-dev/october_backend/pkg/logger"
)

type CompanyHandler struct {
	service company.Service
	logger  logger.Logger
}

func NewCompanyHandler(service company.Service, logger logger.Logger) *CompanyHandler {
	return &CompanyHandler{
		service: service,
		logger:  logger,
	}
}

// GET /company/{company-name}
func (h *CompanyHandler) GetCompanyByName(w http.ResponseWriter, r *http.Request) {
	// Extract company name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/company/")
	companyName := strings.TrimSpace(path)

	if companyName == "" {
		h.logger.Warn("Empty company name in request", "path", r.URL.Path)
		h.writeErrorResponse(w, http.StatusBadRequest, "company name is required")
		return
	}

	// URL decode the company name
	companyName = strings.ReplaceAll(companyName, "%20", " ")
	companyName = strings.ReplaceAll(companyName, "+", " ")

	h.logger.Info("Getting company by name", "name", companyName, "client_ip", utils.GetClientIP(r))

	// Get company from service
	companyResp, err := h.service.GetCompanyByName(r.Context(), companyName)
	if err != nil {
		if err == company.ErrCompanyNotFound {
			h.logger.Info("Company not found", "name", companyName)
			h.writeErrorResponse(w, http.StatusNotFound, "company not found")
			return
		}

		h.logger.Error("Failed to get company", "error", err, "name", companyName)
		h.writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, companyResp)
}

// GET /companies - Get all companies with optional pagination
func (h *CompanyHandler) GetAllCompanies(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting all companies", "client_ip", utils.GetClientIP(r))

	// Parse pagination parameters
	query := r.URL.Query()
	limit := 20  // default limit
	offset := 0  // default offset

	if limitStr := query.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			if parsedLimit > 100 {
				parsedLimit = 100 // max limit
			}
			limit = parsedLimit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get companies from service
	companies, err := h.service.ListCompanies(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get all companies", "error", err, "limit", limit, "offset", offset)
		h.writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Create response with pagination info
	response := map[string]interface{}{
		"companies": companies,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"count":  len(companies),
		},
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *CompanyHandler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var req company.CreateCompanyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid JSON in create company request", "error", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	h.logger.Info("Creating company", "name", req.Name, "ticker", req.Ticker)

	companyResp, err := h.service.CreateCompany(r.Context(), &req)
	if err != nil {
		if err == company.ErrCompanyExists {
			h.logger.Info("Company already exists", "name", req.Name)
			h.writeErrorResponse(w, http.StatusConflict, "company already exists")
			return
		}

		if err == company.ErrInvalidCompanyData || strings.Contains(err.Error(), "invalid") {
			h.logger.Warn("Invalid company data", "error", err)
			h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		h.logger.Error("Failed to create company", "error", err, "name", req.Name)
		h.writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, companyResp)
}

func (h *CompanyHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
	}
}

func (h *CompanyHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResp := map[string]interface{}{
		"error":   true,
		"message": message,
		"status":  statusCode,
	}

	h.writeJSONResponse(w, statusCode, errorResp)
}