package handler

import (
	"fmt"
	"frappuccino/internal/service"
	"frappuccino/utils"
	"log/slog"
	"net/http"
	"time"
)

type ReportHandlerInterface interface {
	HandleGetTotalSales(w http.ResponseWriter, r *http.Request)
	// HandleGetPopulatItem(w http.ResponseWriter, r *http.Request)
	HandleGetOrderedItemsByPeriod(w http.ResponseWriter, r *http.Request)
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

func (h ReportHandler) HandleGetOrderedItemsByPeriod(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	month := r.URL.Query().Get("month")
	year := r.URL.Query().Get("year")

	if period == "" {
		http.Error(w, "Period is required", http.StatusBadRequest)
		return
	}

	if period == "day" {
		if month == "" {
			month = time.Now().Month().String()
		}
		if year == "" {
			year = fmt.Sprintf("%d", time.Now().Year())
		}
	} else if period == "month" {
		if year == "" {
			year = fmt.Sprintf("%d", time.Now().Year())
		}
	}

	orderedItems, err := h.reportService.GetOrderedItemsByPeriod(period, month, year)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve ordered items: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"period":       period,
		"month":        month,
		"year":         year,
		"orderedItems": orderedItems,
	}

	utils.ResponseInJSON(w, http.StatusOK, response)
}
