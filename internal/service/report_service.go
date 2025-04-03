package service

import (
	"errors"
	"fmt"
	"log/slog"

	"frappuccino/internal/dal"
	"frappuccino/models"
)

type ReportServiceInterface interface {
	GetTotalSales() (float64, error)
	GetPopularItems() ([]models.MenuItem, error)
	GetOrderedItemsByPeriod(period string, month string, year int) ([]models.OrderItemReport, error)
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

func (s *ReportService) GetOrderedItemsByPeriod(period string, month string, year int) ([]models.OrderItemReport, error) {
	if period == "day" {
		result, err := s.reportRepo.GetOrderedItemsByDay(month)
		if err != nil {
			return nil, err
		}
		return result, nil
	} else if period == "month" {
		result, err := s.reportRepo.GetOrderedItemsByMonth(year)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	return nil, errors.New("неверное значение параметра period")
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
