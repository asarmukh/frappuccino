package dal

import (
	"database/sql"
	"errors"
	"frappuccino/models"
	"log"
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
	// Валидация здесь не нужна вынесу в сервисы потом
	if order.CustomerName == "" {
		return models.Order{}, errors.New("customer name cannot be empty")
	}
	if order.TotalAmount < 0 {
		return models.Order{}, errors.New("total amount cannot be negative")
	}

	// Находим или создаем клиента по имени
	var customerID int
	customerErr := r.db.QueryRow(
		"SELECT id FROM customers WHERE name = $1",
		order.CustomerName,
	).Scan(&customerID)

	if customerErr == sql.ErrNoRows {
		// Клиент не существует, создаем нового
		customerErr = r.db.QueryRow(
			"INSERT INTO customers (name) VALUES ($1) RETURNING id",
			order.CustomerName,
		).Scan(&customerID)

		if customerErr != nil {
			log.Printf("Customer creation error: %v", customerErr)
			return models.Order{}, customerErr
		}
	} else if customerErr != nil {
		log.Printf("Customer search error: %v", customerErr)
		return models.Order{}, customerErr
	}

	// Создаем заказ
	var newOrder models.Order

	query := `INSERT INTO orders
			(customer_id, total_amount, special_instructions, payment_method, is_completed)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, status, payment_method, created_at`

	err := r.db.QueryRow(
		query,
		customerID,
		order.TotalAmount,
		sql.NullString{String: order.SpecialInstructions, Valid: order.SpecialInstructions != ""},
		order.PaymentMethod,
		order.IsCompleted,
	).Scan(&newOrder.ID, &newOrder.Status, &newOrder.PaymentMethod, &newOrder.CreatedAt)
	if err != nil {
		log.Printf("Error inserting orderа: %v", err)
		return models.Order{}, err
	}

	// Сохраняем Items в модели, если они есть
	if len(order.Items) > 0 {
		newOrder.Items = order.Items
	}

	// Присваиваем остальные поля
	newOrder.CustomerName = order.CustomerName
	newOrder.TotalAmount = order.TotalAmount
	newOrder.SpecialInstructions = order.SpecialInstructions
	newOrder.PaymentMethod = order.PaymentMethod
	newOrder.IsCompleted = order.IsCompleted

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
