package dal

import (
	"database/sql"
	"encoding/json"
	"frappuccino/models"
	"log"
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

// Method for adding a new order to the database
func (r *OrderRepository) AddOrder(order models.Order) (models.Order, error) {
	var newOrder models.Order
	query := `INSERT INTO orders
			(name, total_amount, special_instructions)
			VALUES ($1, $2, $3)
			RETURNING id, status, total_amount, created_at`

	specialInstrustionByte, err1 := json.Marshal(order.SpecialInstructions)
	if err1 != nil {
		return models.Order{}, err1
	}

	err := r.db.QueryRow(
		query,
		order.CustomerName,
		order.TotalAmount,
		specialInstrustionByte,
	).Scan(&newOrder.ID, &newOrder.CustomerName, &newOrder.Status, &newOrder.TotalAmount, &newOrder.SpecialInstructions, &newOrder.IsCompleted, &newOrder.CreatedAt, &newOrder.UpdatedAt)
	if err != nil {
		log.Printf("Error inserting order: %v", err)
		return models.Order{}, err
	}

	// Если есть items, то добавляем их
	if len(order.Items) > 0 {
		newOrder.Items = order.Items
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
