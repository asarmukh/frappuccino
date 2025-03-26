package dal

import (
	"database/sql"
	"encoding/json"
	"frappuccino/models"
	"log"
)

type OrderRepositoryInterface interface {
	AddOrder(order models.Order) (models.Order, error)
	// LoadOrders() ([]models.Order, error)
	// SaveOrders(orders []models.Order) error
}

type OrderRepository struct {
	db *sql.DB
}

func NewOrderPostgresRepository(db *sql.DB) OrderRepository {
	return OrderRepository{db: db}
}

// Method for adding a new order to the database
func (r OrderRepository) AddOrder(order models.Order) (models.Order, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return models.Order{}, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			log.Println("Transaction rolled back:", err)
		} else {
			err = tx.Commit()
			if err != nil {
				log.Println("Commit error:", err)
			}
		}
	}()

	query := `INSERT INTO orders (name, total_amount, special_instructions)
			VALUES ($1, $2, $3)
			RETURNING id, name, status, total_amount, special_instructions, created_at, updated_at`

	specialInstructionsByte, err := json.Marshal(order.SpecialInstructions)
	if err != nil {
		return models.Order{}, err
	}
	var specialInstructionsData []byte
	if err := r.db.QueryRow(
		query,
		order.CustomerName,
		order.TotalAmount,
		specialInstructionsByte,
	).Scan(&order.ID, &order.CustomerName, &order.Status, &order.TotalAmount, &specialInstructionsData, &order.CreatedAt, &order.UpdatedAt); err != nil {
		log.Printf("Error inserting order: %v", err)
		return models.Order{}, err
	}
	// Декодируем JSONB в map[string]string
	if err := json.Unmarshal(specialInstructionsData, &order.SpecialInstructions); err != nil {
		return models.Order{}, err
	}

	queryPrice := `SELECT price FROM menu_items WHERE id = $1`
	for _, product := range order.Items {
		err := r.db.QueryRow(queryPrice, product.ProductID).Scan(&product.Price)
		if err != nil {
			return models.Order{}, err
		}

		query := `INSERT INTO order_items (order_id, menu_item_id, quantity, price)
	          VALUES ($1, $2, $3, $4)`
		_, err = tx.Exec(query, order.ID, product.ProductID, product.Quantity, product.Price)
		if err != nil {
			return models.Order{}, err
		}
	}

	return order, nil
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
