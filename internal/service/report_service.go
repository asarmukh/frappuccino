package service

import (
	"fmt"
	"frappuccino/internal/dal"
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

// func (r ReportService) GetPopularItems() ([]models.MenuItem, error) {
// 	orders, err := r.orderSevice.GetAllOrders()
// 	if err != nil {
// 		return []models.MenuItem{}, fmt.Errorf("Failed to retrieve orders")
// 	}

// 	orderMap := make(map[string]int)
// 	for _, order := range orders {
// 		for _, item := range order.Items {
// 			orderMap[item.ProductID] += item.Quantity
// 		}
// 	}

// 	menuItems, err := r.menuService.GetAllMenuItems()
// 	if err != nil {
// 		return []models.MenuItem{}, fmt.Errorf("Failed to retrieve menu items")
// 	}

// 	menuMap := make(map[string]models.MenuItem)
// 	for _, item := range menuItems {
// 		menuMap[item.ID] = item
// 	}

// 	popularItems := []models.MenuItem{}

// 	for id, count := range orderMap {
// 		if item, exists := menuMap[id]; exists {
// 			item.Price = float64(count)
// 			popularItems = append(popularItems, item)
// 		}
// 	}

// 	sort.Slice(popularItems, func(i, j int) bool {
// 		return popularItems[i].Price > popularItems[j].Price
// 	})

// 	return popularItems, nil
// }
