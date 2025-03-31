package service

import (
	"fmt"
	"frappuccino/internal/dal"
	"frappuccino/models"
)

type ReportServiceInterface interface {
	GetTotalSales() (float64, error)
}

type ReportService struct {
	// menuService MenuService
	reportRepo dal.ReportRepository
}

func NewReportService(_reportRepo dal.ReportRepository) ReportService {
	return ReportService{
		reportRepo: _reportRepo,
	}
}

func (r ReportService) GetTotalSales() (float64, error) {
	totalSales, err := r.reportRepo.TotalSales()
	if err != nil {
		return -1, fmt.Errorf("error getting total sales")
	}
	return totalSales, nil
}

func (s ReportService) GetPopularItems() ([]models.MenuItem, error) {
	return s.reportRepo.GetPopularItems()
}
