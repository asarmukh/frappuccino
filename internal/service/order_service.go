package service

import (
	"fmt"
	"frappuccino/internal/dal"
	"frappuccino/models"
	"frappuccino/utils"
	"log"
)

type OrderServiceInterface interface {
	CreateOrder(order models.Order) (models.Order, error)
	GetAllOrders() ([]models.Order, error)
	GetOrderByID(id string) (models.Order, error)
	DeleteOrder(id string) (models.Order, error)
	UpdateOrder(id string) (models.Order, error)
	CloseOrder(orderID string) (models.Order, error)
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
	totalAmount := 0.0
	for i, product := range order.Items {
		price, err := s.menuRepo.GetProductPrice(product.ProductID)
		if err != nil {
			return models.Order{}, err
		}
		order.Items[i].Price = price
		totalAmount += price * float64(product.Quantity)
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

// func (s OrderService) DeleteOrder(id string) error {
// 	orders, err := s.GetAllOrders()
// 	if err != nil {
// 		return fmt.Errorf("failed to delete order with ID %s: %v", id, err)
// 	}

// 	indexToDelete := -1
// 	for i, order := range orders {
// 		if order.ID == id {
// 			indexToDelete = i
// 			break
// 		}
// 	}

// 	if indexToDelete == -1 {
// 		return fmt.Errorf("order with ID %s not found", id)
// 	}

// 	orders = append(orders[:indexToDelete], orders[indexToDelete+1:]...)

// 	if err := s.repository.SaveOrders(orders); err != nil {
// 		return fmt.Errorf("could not save orders")
// 	}
// 	return nil
// }

// func (s OrderService) UpdateOrder(id string, changeOrder models.Order) (models.Order, error) {
// 	if changeOrder.CustomerName == "" || changeOrder.Items == nil {
// 		return models.Order{}, errors.New("invalid request body")
// 	}
// 	if changeOrder.ID != "" {
// 		return models.Order{}, fmt.Errorf("cannot change ID or add in body requsest")
// 	}

// 	orders, err := s.repository.LoadOrders()
// 	if err != nil {
// 		return changeOrder, fmt.Errorf("error reading all oreders %s: %v", id, err)
// 	}

// 	menu, err := s.menuService.repository.LoadMenuItems()
// 	if err != nil {
// 		return models.Order{}, err
// 	}

// 	err = utils.ValidateOrder(menu, changeOrder)
// 	if err != nil {
// 		return models.Order{}, err
// 	}

// 	for i := 0; i < len(orders); i++ {
// 		if orders[i].ID == changeOrder.ID && changeOrder.Status == "closed" {
// 			return models.Order{}, fmt.Errorf("order is closed")
// 		}

// 		if orders[i].ID == id {
// 			orders[i].CustomerName = changeOrder.CustomerName
// 			orders[i].CreatedAt = time.Now().UTC().Format(time.RFC3339)
// 			orders[i].Items = changeOrder.Items
// 			s.repository.SaveOrders(orders)

// 			return orders[i], nil
// 		}
// 	}
// 	return changeOrder, fmt.Errorf("order with ID %s not found", id)
// }

// func (s OrderService) CloseOrder(id string) (models.Order, error) {
// 	orders, err := s.repository.LoadOrders()
// 	if err != nil {
// 		return models.Order{}, fmt.Errorf("order with ID %s not found", id)
// 	}

// 	orderId, err := s.GetOrderByID(id)
// 	if err != nil {
// 		return models.Order{}, fmt.Errorf("failed to retrieve order by ID%s", id)
// 	}

// 	if orderId.Status == "closed" {
// 		return models.Order{}, fmt.Errorf("opration not allowed")
// 	}

// 	menu, err := s.menuService.repository.LoadMenuItems()
// 	if err != nil {
// 		return models.Order{}, err
// 	}

// 	menuMap := make(map[string]models.MenuItem)
// 	for _, items := range menu {
// 		menuMap[items.ID] = items
// 	}

// 	inventory, err := s.inventoryService.GetAllInventory()
// 	if err != nil {
// 		return models.Order{}, fmt.Errorf("failed to retrieve inventory")
// 	}

// 	var newDataMenu []models.MenuItem

// 	ingredientMap := make(map[string]models.InventoryItem)
// 	for _, items := range inventory {
// 		ingredientMap[items.IngredientID] = items
// 	}

// 	for _, items := range orderId.Items {
// 		for i := 0; i < items.Quantity; i++ {
// 			if item, exists := menuMap[items.ProductID]; exists {
// 				newDataMenu = append(newDataMenu, item)
// 			}
// 		}
// 		if err := utils.ValidateQuantity(float64(items.Quantity)); err != nil {
// 			return models.Order{}, err
// 		}
// 	}

// 	for _, items := range newDataMenu {
// 		for _, ingredient := range items.Ingredients {
// 			if item, exist := ingredientMap[ingredient.IngredientID]; exist {
// 				fmt.Printf("Checking ingredient ID: %v, required: %v, available: %v\n",
// 					ingredient.IngredientID, ingredient.Quantity, item.Quantity)
// 				if ingredient.Quantity > item.Quantity {
// 					return models.Order{}, fmt.Errorf("not enough quantity for ingredient ID %v: required %v, available %v", ingredient.IngredientID, ingredient.Quantity, item.Quantity)
// 				}
// 				item.Quantity -= ingredient.Quantity
// 				ingredientMap[ingredient.IngredientID] = item
// 			}
// 		}
// 	}

// 	for ingredientID, item := range ingredientMap {
// 		if _, err := s.inventoryService.UpdateInventoryItem(ingredientID, item); err != nil {
// 			return models.Order{}, fmt.Errorf("failed to update inventory for ingredientID %v", ingredientID)
// 		}
// 	}

// 	for i := 0; i < len(orders); i++ {
// 		if orders[i].ID == id {
// 			orders[i].Status = "closed"
// 			s.repository.SaveOrders(orders)
// 			return orders[i], nil
// 		}
// 	}
// 	return models.Order{}, fmt.Errorf("order with ID %s not found", id)
// }
