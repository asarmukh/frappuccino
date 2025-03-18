package handler

import (
	"net/http"

	"frappuccino/internal/service"
)

type ReportHandlerInterface interface {
	HandleGetTotalSales(w http.ResponseWriter, r *http.Request)
	HandleGetPopulatItem(w http.ResponseWriter, r *http.Request)
}

type ReportHandler struct {
	reportService service.ReportService
}

func NewReportHandler(reportService service.ReportService) ReportHandler {
	return ReportHandler{reportService: reportService}
}

// func (h ReportHandler) HandleGetTotalSales(w http.ResponseWriter, r *http.Request) {
// 	slog.Info("Received request to get total sales")

// 	totalSales, err := h.reportService.GetTotalSales()
// 	if err != nil {
// 		slog.Error("Error fetching total sales", "error", err)
// 		http.Error(w, "Failed to retrieve total sales", http.StatusInternalServerError)
// 		return
// 	}

// 	response := map[string]float64{"total sales": totalSales}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(response)

// 	slog.Info("Total sales response sent successfully")
// }

// func (h ReportHandler) HandleGetPopulatItem(w http.ResponseWriter, r *http.Request) {
// 	slog.Info("Received request to get popular items")

// 	popularItem, err := h.reportService.GetPopularItems()
// 	if err != nil {
// 		slog.Error("Error fetching popular item", "error", err)
// 		http.Error(w, "Failed to retrieve popular item", http.StatusInternalServerError)
// 	}

// 	utils.ResponseInJSON(w, 200, popularItem)

// 	slog.Info("Popular item response sent successfully")
// }
