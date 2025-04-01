package service

import (
	"fmt"
	"frappuccino/internal/dal"
	"frappuccino/models"
	"frappuccino/utils"
	"log"
	"log/slog"
	"time"
)

type OrderServiceInterface interface {
	CreateOrder(order models.Order) (models.Order, error)
	GetAllOrders() ([]models.Order, error)
	GetOrderByID(id int) (models.Order, error)
	DeleteOrder(id int) error
	UpdateOrder(id int) (models.Order, error)
	CloseOrder(id int) (models.Order, error)
	GetNumberOfOrderedItems(startDate, endDate string) (map[string]int, error)
}

type OrderService struct {
	orderRepo dal.OrderRepository
	menuRepo  dal.MenuRepository
}

func NewOrderService(_orderRepo dal.OrderRepository, _menuRepo dal.MenuRepository) OrderService {
	return OrderService{
		orderRepo: _orderRepo,
		menuRepo:  _menuRepo,
	}
}

// Method of creating a new order
func (s OrderService) CreateOrder(order models.Order) (models.Order, error) {
	if err := utils.IsValidName(order.CustomerName); err != nil {
		return models.Order{}, err
	}

	if err := utils.ValidateSpecialInstructions(order.SpecialInstructions); err != nil {
		return models.Order{}, err
	}

	// Checking that all products exist on the menu
	for _, product := range order.Items {
		exists, err := s.menuRepo.ProductExists(product.ProductID)
		if err != nil {
			return models.Order{}, err
		}
		if !exists {
			return models.Order{}, fmt.Errorf("product with ID %d not found", product.ProductID)
		}
	}

	// Calculating the total amount of the order
	totalAmount, err := s.TotalAmount(order)
	if err != nil {
		return models.Order{}, err
	}
	order.TotalAmount = totalAmount

	newOrder, err := s.orderRepo.AddOrder(order)
	if err != nil {
		return models.Order{}, fmt.Errorf("error creating order: %w", err)
	}

	log.Printf("order added: %d", newOrder.ID)
	return newOrder, nil
}

func (s OrderService) GetAllOrders() ([]models.Order, error) {
	orders, err := s.orderRepo.LoadOrders()
	if err != nil {
		log.Printf("error get all orders!")
		return nil, err
	}
	return orders, nil
}

func (s OrderService) GetOrderByID(id int) (models.Order, error) {
	order, err := s.orderRepo.LoadOrder(id)
	if err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func (s OrderService) DeleteOrder(id int) error {
	err := s.orderRepo.DeleteOrderByID(id)
	if err != nil {
		slog.Warn("Failed to delete order", "orderID", id, "error", err)
		return fmt.Errorf("failed to delete order with ID %d: %v", id, err)
	}

	return nil
}

func (s OrderService) UpdateOrder(id int, changeOrder models.Order) (models.Order, error) {
	if err := utils.IsValidName(changeOrder.CustomerName); err != nil {
		return models.Order{}, err
	}

	if err := utils.ValidateSpecialInstructions(changeOrder.SpecialInstructions); err != nil {
		return models.Order{}, err
	}

	// Checking that all products exist on the menu
	for _, product := range changeOrder.Items {
		exists, err := s.menuRepo.ProductExists(product.ProductID)
		if err != nil {
			return models.Order{}, err
		}
		if !exists {
			return models.Order{}, fmt.Errorf("product with ID %d not found", product.ProductID)
		}
	}

	// Calculating the total amount of the order
	totalAmount, err := s.TotalAmount(changeOrder)
	if err != nil {
		return models.Order{}, err
	}
	changeOrder.TotalAmount = totalAmount

	order, err := s.orderRepo.UpdateOrder(id, changeOrder)
	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func (s OrderService) CloseOrder(id int) (models.Order, error) {
	order, err := s.orderRepo.CloseOrder(id)
	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func (s OrderService) GetNumberOfOrderedItems(startDate, endDate string) (map[string]int, error) {
	var start time.Time
	var end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, fmt.Errorf("invalid startDate format, expected YYYY-MM-DD: %w", err)
		}
	} else {
		start = time.Time{}
	}

	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, fmt.Errorf("invalid endDate format, expected YYYY-MM-DD: %w", err)
		}
	} else {
		end = time.Now()
	}

	orderedItems, err := s.orderRepo.GetOrderedItemsCount(start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ordered items: %w", err)
	}

	return orderedItems, nil
}

func (s OrderService) TotalAmount(order models.Order) (float64, error) {
	// Calculating the total amount of the order
	totalAmount := 0.0
	for i, product := range order.Items {
		price, err := s.menuRepo.GetProductPrice(product.ProductID)
		if err != nil {
			return 0.0, err
		}
		order.Items[i].Price = price
		totalAmount += price * float64(product.Quantity)
	}
	return totalAmount, nil
}
