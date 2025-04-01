package handler

import (
	"fmt"
	"frappuccino/internal/service"
	"frappuccino/utils"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type ReportHandlerInterface interface {
	HandleGetTotalSales(w http.ResponseWriter, r *http.Request)
	HandleGetPopularItems(w http.ResponseWriter, r *http.Request)
	HandleSearch(w http.ResponseWriter, r *http.Request)
}

type ReportHandler struct {
	reportService service.ReportService
}

func NewReportHandler(reportService service.ReportService) ReportHandler {
	return ReportHandler{reportService: reportService}
}

func (h ReportHandler) HandleGetTotalSales(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request to get total sales")

	totalSales, err := h.reportService.GetTotalSales()
	if err != nil {
		slog.Error("Failed to fetch total sales from service", "error", err.Error())
		http.Error(w, "Failed to retrieve total sales", http.StatusInternalServerError)
		return
	}

	response := map[string]float64{"total_sales": totalSales}

	utils.ResponseInJSON(w, http.StatusOK, response)
	slog.Info("✅ Total sales response sent successfully", "total_sales", totalSales)
}

func (h ReportHandler) HandleGetPopularItems(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request to get popular items")

	popularItems, err := h.reportService.GetPopularItems()
	if err != nil {
		slog.Error("Error fetching popular items", "error", err)
		http.Error(w, "Failed to retrieve popular items", http.StatusInternalServerError)
		return
	}

	utils.ResponseInJSON(w, http.StatusOK, popularItems)
	slog.Info("Popular items response sent successfully")
}

func (h ReportHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request to get search")

	q := r.URL.Query().Get("q")
	filterParam := r.URL.Query().Get("filter")
	minPriceStr := r.URL.Query().Get("minPrice")
	maxPriceStr := r.URL.Query().Get("maxPrice")

	var filters []string
	if filterParam != "" {
		filters = strings.Split(filterParam, ",")
	}

	if q == "" {
		utils.ErrorInJSON(w, http.StatusBadRequest, fmt.Errorf("search query string required!"))
		return
	}

	minPrice, err := strconv.Atoi(minPriceStr)
	if err != nil {
		minPrice = 0
	}

	maxPrice, err := strconv.Atoi(maxPriceStr)
	if err != nil {
		maxPrice = 100000
	}

	slog.Info("Min price: %d; Max price: %d; q: %s; filters: %v", minPrice, maxPrice, q, filters)

	searchResult, err := h.reportService.Search(q, filters, minPrice, maxPrice)
	if err != nil {
		slog.Error("Error fetching search", "error", err)
		utils.ErrorInJSON(w, http.StatusInternalServerError, err)
		return
	}

	utils.ResponseInJSON(w, http.StatusOK, searchResult)
	slog.Info("Search response sent successfully")
}
