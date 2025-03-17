package dal

import (
	"database/sql"
	"errors"
	"log"

	"frappuccino/models"
)

type OrderRepositoryInterface interface {
	AddOrder(order models.Order) (models.Order, error)
	LoadOrders() ([]models.Order, error)
	SaveOrders(orders []models.Order) error
}

type OrderPostgresRepository struct {
	db *sql.DB
}

func NewOrderPostgresRepository(db *sql.DB) *OrderPostgresRepository {
	return &OrderPostgresRepository{db: db}
}

func (r *OrderPostgresRepository) AddOrder(order models.Order) (models.Order, error) {
	if order.CustomerName == "" {
		return models.Order{}, errors.New("customer name cannot be empty")
	}
	if order.TotalAmount < 0 {
		return models.Order{}, errors.New("total amount cannot be negative")
	}
	if order.PaymentMethod == "" {
		return models.Order{}, errors.New("payment method is required")
	}

	query := `INSERT INTO orders 
        (customer_name, status, total_amount, special_instructions, payment_method, is_completed, created_at, updated_at) 
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) 
        RETURNING id, created_at, updated_at`

	var newOrder models.Order
	err := r.db.QueryRow(
		query,
		order.CustomerName,
		order.Status,
		order.TotalAmount,
		sql.NullString{String: order.SpecialInstructions, Valid: order.SpecialInstructions != ""},
		order.PaymentMethod,
		order.IsCompleted,
	).Scan(&newOrder.ID, &newOrder.CreatedAt, &newOrder.UpdatedAt)

	if err != nil {
		log.Printf("Ошибка вставки заказа: %v", err)
		return models.Order{}, err
	}

	// Присваиваем остальные поля
	newOrder.CustomerName = order.CustomerName
	newOrder.Status = order.Status
	newOrder.TotalAmount = order.TotalAmount
	newOrder.SpecialInstructions = order.SpecialInstructions
	newOrder.PaymentMethod = order.PaymentMethod
	newOrder.IsCompleted = order.IsCompleted

	log.Printf("Заказ успешно добавлен: ID %s", newOrder.ID)
	return newOrder, nil
}

// func (r OrderRepositoryJSON) LoadOrders() ([]models.Order, error) {
// 	filePath := filepath.Join(r.filePath, "orders.json")
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	var orders []models.Order
// 	if err := json.NewDecoder(file).Decode(&orders); err != nil {
// 		return nil, err
// 	}

// 	return orders, nil
// }

// func (r OrderRepositoryJSON) SaveOrders(orders []models.Order) error {
// 	filePath := filepath.Join(r.filePath, "orders.json")
// 	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
// 	if err != nil {
// 		return fmt.Errorf("could not open or create inventory file: %v", err)
// 	}
// 	defer file.Close()

// 	encoder := json.NewEncoder(file)
// 	encoder.SetIndent("", "  ")
// 	if err := encoder.Encode(orders); err != nil {
// 		return fmt.Errorf("could not encode inventory to file: %v", err)
// 	}

// 	return nil
// }
