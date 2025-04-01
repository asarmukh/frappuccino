package service

import (
	"fmt"
	"frappuccino/internal/dal"
	"frappuccino/models"
)

type ReportServiceInterface interface {
	GetTotalSales() (float64, error)
	GetOrderedItemsByPeriod(period, month, year string) (interface{}, error)
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

func (s ReportService) GetOrderedItemsByPeriod(period, month, year string) (interface{}, error) {
	var orderedItems interface{}
	var err error

	if period == "day" {
		orderedItems, err = s.reportRepo.GetOrderedItemsByDay(month, year)
	} else if period == "month" {
		orderedItems, err = s.reportRepo.GetOrderedItemsByMonth(year)
	} else {
		return nil, fmt.Errorf("invalid period parameter")
	}

	if err != nil {
		return nil, fmt.Errorf("error getting ordered items by period: %w", err)
	}

	return orderedItems, nil
}
