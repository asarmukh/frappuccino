package dal

import (
	"database/sql"
	"fmt"
	"frappuccino/models"
)

type OrderRepositoryInterface interface {
	AddOrder(order models.Order) (models.Order, error)
	LoadOrders() ([]models.Order, error)
	SaveOrders(orders []models.Order) error
}

type OrderRepository struct {
	db *sql.DB
}

func NewOrderPostgresRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) AddOrder(customerID int, totalAmount float64, specialInstructions, paymentMethod string, isCompleted bool) (models.Order, error) {
	var newOrder models.Order
	query := `INSERT INTO orders
              (customer_id, total_amount, special_instructions, payment_method, is_completed)
              VALUES ($1, $2, $3, $4, $5)
              RETURNING id, status, payment_method, created_at`

	err := r.db.QueryRow(query, customerID, totalAmount,
		sql.NullString{String: specialInstructions, Valid: specialInstructions != ""},
		paymentMethod, isCompleted).Scan(&newOrder.ID, &newOrder.Status, &newOrder.PaymentMethod, &newOrder.CreatedAt)
	if err != nil {
		return models.Order{}, fmt.Errorf("error inserting order: %w", err)
	}
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
