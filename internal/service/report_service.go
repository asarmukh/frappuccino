package service

import (
	"fmt"
	"frappuccino/internal/dal"
	"frappuccino/models"
	"log/slog"
)

type ReportServiceInterface interface {
	GetTotalSales() (float64, error)
	GetPopularItems() ([]models.MenuItem, error)
	Search(q string, filters []string, minPrice int, maxPrice int) (models.SearchResult, error)
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

func (s ReportService) Search(q string, filters []string, minPrice int, maxPrice int) (models.SearchResult, error) {
	var searchResult models.SearchResult
	var hadError bool

	// Если фильтры не заданы, ищем везде
	if len(filters) == 0 {
		filters = []string{"menu", "orders"}
	}

	for _, filter := range filters {
		switch filter {
		case "menu":
			menuItems, err := s.reportRepo.SearchMenu(q, minPrice, maxPrice)
			if err != nil {
				slog.Error("Error searching menu", "error", err)
				hadError = true
				continue
			}
			searchResult.MenuItems = append(searchResult.MenuItems, menuItems...)

		case "orders":
			// Заглушка для orders, пока нет реализации
			slog.Info("Orders search not implemented yet")
		}
	}

	if hadError {
		return searchResult, fmt.Errorf("some errors occurred during search, check logs")
	}

	return searchResult, nil
}
