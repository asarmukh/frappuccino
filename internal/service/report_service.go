package service

import (
	"fmt"
	"frappuccino/internal/dal"
	"frappuccino/models"
	"log/slog"
)

type ReportServiceInterface interface {
	GetTotalSales() (float64, error)
	GetOrderedItemsByPeriod(period, month, year string) (interface{}, error)
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

func (s ReportService) Search(q string, filters []string, minPrice int, maxPrice int) (models.SearchResult, error) {
	slog.Info("Start search...", "query", q, "filters", filters, "minPrice", minPrice, "maxPrice", maxPrice)
	var searchResult models.SearchResult
	var hadError bool

	// Если фильтры не заданы, ищем везде
	if len(filters) == 0 {
		filters = []string{"menu", "orders"}
		slog.Info("No filters provided, searching in all categories")
	}

	for _, filter := range filters {
		slog.Info("Processing filter", "filter", filter)

		switch filter {
		case "menu":
			slog.Info("Searching in menu...")
			searchMenu, err := s.reportRepo.SearchMenu(q, minPrice, maxPrice)
			if err != nil {
				slog.Error("Error searching menu", "error", err)
				hadError = true
				continue
			}
			slog.Info("Menu search results", "totalMatches", searchMenu.TotalMatches, "items", len(searchMenu.MenuItems))
			searchResult.TotalMatches += searchMenu.TotalMatches
			searchResult.MenuItems = searchMenu.MenuItems

		case "orders":
			slog.Info("Searching in orders...")
			searchOrders, err := s.reportRepo.SearchOrders(q, minPrice, maxPrice)
			if err != nil {
				slog.Error("Error searching orders", "error", err)
				hadError = true
				continue
			}
			slog.Info("Orders search results", "totalMatches", searchOrders.TotalMatches, "orders", len(searchOrders.Orders))
			searchResult.TotalMatches += searchOrders.TotalMatches
			searchResult.Orders = searchOrders.Orders
		default:
			slog.Warn("Unknown filter, skipping", "filter", filter)
		}
	}

	if hadError {
		slog.Error("Search completed with errors, check previous logs")
		return searchResult, fmt.Errorf("some errors occurred during search, check logs")
	}

	slog.Info("Search completed successfully", "totalMatches", searchResult.TotalMatches)
	return searchResult, nil
}
