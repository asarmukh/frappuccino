package service

import (
	"fmt"
	"sort"

	"frappuccino/models"
)

type ReportService struct {
	menuService MenuService
	orderSevice OrderService
}

func NewReportService(menuService MenuService, orderService OrderService) ReportService {
	return ReportService{
		menuService: menuService,
		orderSevice: orderService,
	}
}

func (r ReportService) GetTotalSales() (float64, error) {
	orders, err := r.orderSevice.GetAllOrders()
	if err != nil {
		return 0, fmt.Errorf("Failed to retrieve orders")
	}

	menuItems, err := r.menuService.GetAllMenuItems()
	if err != nil {
		return 0, fmt.Errorf("Failed to retrieve menu items")
	}

	totalSales := 0.0

	for _, order := range orders {
		for _, item := range order.Items {
			for _, menuItem := range menuItems {
				if item.ProductID == menuItem.ID {
					fmt.Println(item.Quantity, menuItem.Price)
					totalSales += float64(item.Quantity) * menuItem.Price
				}
			}
		}
	}

	return totalSales, nil
}

func (r ReportService) GetPopularItems() ([]models.MenuItem, error) {
	orders, err := r.orderSevice.GetAllOrders()
	if err != nil {
		return []models.MenuItem{}, fmt.Errorf("Failed to retrieve orders")
	}

	orderMap := make(map[string]int)
	for _, order := range orders {
		for _, item := range order.Items {
			orderMap[item.ProductID] += item.Quantity
		}
	}

	menuItems, err := r.menuService.GetAllMenuItems()
	if err != nil {
		return []models.MenuItem{}, fmt.Errorf("Failed to retrieve menu items")
	}

	menuMap := make(map[string]models.MenuItem)
	for _, item := range menuItems {
		menuMap[item.ID] = item
	}

	popularItems := []models.MenuItem{}

	for id, count := range orderMap {
		if item, exists := menuMap[id]; exists {
			item.Price = float64(count)
			popularItems = append(popularItems, item)
		}
	}

	sort.Slice(popularItems, func(i, j int) bool {
		return popularItems[i].Price > popularItems[j].Price
	})

	return popularItems, nil
}
