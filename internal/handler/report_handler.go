package handler

import (
	"log/slog"
	"net/http"

	"frappuccino/internal/service"
	"frappuccino/utils"
)

type ReportHandlerInterface interface {
	HandleGetTotalSales(w http.ResponseWriter, r *http.Request)
	// HandleGetPopulatItem(w http.ResponseWriter, r *http.Request)
}

type ReportHandler struct {
	reportService service.ReportService
}

func NewReportHandler(reportService service.ReportService) ReportHandler {
	return ReportHandler{reportService: reportService}
}

func (h ReportHandler) HandleGetTotalSales(w http.ResponseWriter, r *http.Request) {
	slog.Info("📊 Received request to get total sales")

	totalSales, err := h.reportService.GetTotalSales()
	if err != nil {
		slog.Error("❌ Failed to fetch total sales from service", "error", err.Error())
		http.Error(w, "Failed to retrieve total sales", http.StatusInternalServerError)
		return
	}

	response := map[string]float64{"total_sales": totalSales}

	utils.ResponseInJSON(w, http.StatusOK, response)
	slog.Info("✅ Total sales response sent successfully", "total_sales", totalSales)
}

// func (h ReportHandler) HandleGetPopularItems(w http.ResponseWriter, r *http.Request) {
// 	slog.Info("Received request to get popular items")

// 	popularItem, err := h.reportService.GetPopularItems()
// 	if err != nil {
// 		slog.Error("Error fetching popular item", "error", err)
// 		http.Error(w, "Failed to retrieve popular item", http.StatusInternalServerError)
// 	}

// 	utils.ResponseInJSON(w, 200, popularItem)

// 	slog.Info("Popular item response sent successfully")
// }
