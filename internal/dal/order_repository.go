package dal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	// "frappuccino/internal/db"
	"frappuccino/models"
	"frappuccino/utils"
	"log"
)

type OrderRepositoryInterface interface {
	AddOrder(order models.Order) (models.Order, error)
	LoadOrders() ([]models.Order, error)
	LoadOrder(id int) (models.Order, error)
	DeleteOrderByID(id int) error
	UpdateOrder(id int) (models.Order, error)
	CloseOrder(id int) (models.Order, error)
}

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
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
	for i := range order.Items {
		err := r.db.QueryRow(queryPrice, order.Items[i].ProductID).Scan(&order.Items[i].Price)
		if err != nil {
			return models.Order{}, err
		}

		query := `INSERT INTO order_items (order_id, menu_item_id, quantity, price)
				  VALUES ($1, $2, $3, $4)`
		_, err = tx.Exec(query, order.ID, order.Items[i].ProductID, order.Items[i].Quantity, order.Items[i].Price)
		if err != nil {
			return models.Order{}, err
		}
	}
	queryItems := `SELECT menu_item_id, quantity, price FROM order_items WHERE order_id = $1`
	rows, err := tx.Query(queryItems, order.ID)
	if err != nil {
		tx.Rollback()
		return models.Order{}, fmt.Errorf("ошибка получения списка товаров заказа: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
			tx.Rollback()
			return models.Order{}, fmt.Errorf("ошибка при сканировании товаров: %w", err)
		}
		items = append(items, item)
	}
	order.Items = items

	return order, nil
}

func (r OrderRepository) LoadOrders() ([]models.Order, error) {
	var orders []models.Order

	// Получаем все заказы
	query := `SELECT id, name, status, total_amount, special_instructions, created_at, updated_at FROM orders`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer rows.Close() // Закрываем `rows` только ОДИН раз

	for rows.Next() {
		var order models.Order
		var specialInstructionsStr string

		// Сканируем данные заказа
		if err := rows.Scan(&order.ID, &order.CustomerName, &order.Status, &order.TotalAmount, &specialInstructionsStr, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %v", err)
		}

		// Декодируем JSON
		if err := json.Unmarshal([]byte(specialInstructionsStr), &order.SpecialInstructions); err != nil {
			return nil, fmt.Errorf("ошибка декодирования JSON: %v", err)
		}

		// Загружаем товары для этого заказа
		queryItems := `SELECT menu_item_id, quantity, price FROM order_items WHERE order_id = $1`
		itemRows, err := r.db.Query(queryItems, order.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения списка товаров заказа: %w", err)
		}
		defer itemRows.Close() // Отдельный `defer` для товаров

		var items []models.OrderItem
		for itemRows.Next() {
			var item models.OrderItem
			if err := itemRows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
				return nil, fmt.Errorf("ошибка при сканировании товаров: %w", err)
			}
			items = append(items, item)
		}
		order.Items = items

		orders = append(orders, order)
	}

	// Проверяем ошибки во внешнем rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации строк: %v", err)
	}

	return orders, nil
}

func (r OrderRepository) LoadOrder(id int) (models.Order, error) {
	var order models.Order
	var specialInstructionsStr string
	query := `SELECT id, name, status, total_amount, special_instructions, created_at, updated_at FROM orders WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&order.ID,
		&order.CustomerName,
		&order.Status,
		&order.TotalAmount,
		&specialInstructionsStr,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Order{}, fmt.Errorf("order with ID %d not found", id)
		}
		return models.Order{}, fmt.Errorf("error getting element: %v", err)
	}
	specialInstructions, err := utils.ConvertSpecialInstructions(specialInstructionsStr)
	if err != nil {
		return models.Order{}, fmt.Errorf("error converting special_instructions: %v", err)
	}

	order.SpecialInstructions = specialInstructions

	// Загружаем товары для этого заказа
	queryItems := `SELECT menu_item_id, quantity, price FROM order_items WHERE order_id = $1`
	itemRows, err := r.db.Query(queryItems, order.ID)
	if err != nil {
		return models.Order{}, fmt.Errorf("ошибка получения списка товаров заказа: %w", err)
	}
	defer itemRows.Close() // Отдельный `defer` для товаров

	var items []models.OrderItem
	for itemRows.Next() {
		var item models.OrderItem
		if err := itemRows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
			return models.Order{}, fmt.Errorf("ошибка при сканировании товаров: %w", err)
		}
		items = append(items, item)
	}
	order.Items = items

	return order, nil
}

func (r OrderRepository) DeleteOrderByID(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transactioOrderRepositoryn: %v", err)
	}
	defer tx.Rollback()

	queryDelete := `DELETE FROM orders WHERE id = $1`
	result, err := tx.Exec(queryDelete, id)
	if err != nil {
		return fmt.Errorf("error while deleting element: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось получить количество затронутых строк: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order with ID %d not found", id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

// Обновление заказа
func (r OrderRepository) UpdateOrder(id int, changeOrder models.Order) (models.Order, error) {
	var orderUpdated models.Order
	var specialInstructionsJSON []byte

	specialInstructionsBytes, err := json.Marshal(changeOrder.SpecialInstructions)
	if err != nil {
		return models.Order{}, fmt.Errorf("JSON marshaling error: %w", err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return models.Order{}, fmt.Errorf("transaction start error: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	queryUpdate := `
        UPDATE orders 
        SET name = $2, status = $3, total_amount = $4, special_instructions = $5::jsonb, updated_at = NOW() 
        WHERE id = $1 
        RETURNING *`
	err = tx.QueryRow(queryUpdate, id, changeOrder.CustomerName, "updated", changeOrder.TotalAmount, specialInstructionsBytes).
		Scan(&orderUpdated.ID, &orderUpdated.CustomerName, &orderUpdated.Status, &orderUpdated.TotalAmount, &specialInstructionsJSON, &orderUpdated.CreatedAt, &orderUpdated.UpdatedAt)
	if err != nil {
		return models.Order{}, fmt.Errorf("request execution error: %w", err)
	}
	// Декодирование JSON обратно в карту
	err = json.Unmarshal(specialInstructionsJSON, &orderUpdated.SpecialInstructions)
	if err != nil {
		return models.Order{}, fmt.Errorf("JSON unmarshaling error: %w", err)
	}

	// Обновление позиций заказа
	queryPrice := `SELECT price FROM menu_items WHERE id = $1`
	queryUpdateItems := `UPDATE order_items SET quantity = $3, price = $4 WHERE order_id = $1 AND menu_item_id = $2`
	for _, item := range changeOrder.Items {
		err := tx.QueryRow(queryPrice, item.ProductID).Scan(&item.Price)
		if err != nil {
			return models.Order{}, fmt.Errorf("error getting a price: %w", err)
		}

		_, err = tx.Exec(queryUpdateItems, orderUpdated.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return models.Order{}, fmt.Errorf("order Item Update Error: %w", err)
		}
	}

	// Получение актуальных позиций заказа
	queryItems := `SELECT menu_item_id, quantity, price FROM order_items WHERE order_id = $1`
	rows, err := tx.Query(queryItems, orderUpdated.ID)
	if err != nil {
		return models.Order{}, fmt.Errorf("error receiving the list of items of the order: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
			return models.Order{}, fmt.Errorf("error when scanning products: %w", err)
		}
		items = append(items, item)
	}
	orderUpdated.Items = items

	// Фиксация транзакции
	if err := tx.Commit(); err != nil {
		return models.Order{}, fmt.Errorf("transaction commit error: %w", err)
	}

	return orderUpdated, nil
}

func (r OrderRepository) CloseOrder(id int) (models.Order, error) {
	order, errLoad := r.LoadOrder(id)
	if errLoad != nil {
		return order, errLoad
	}
	if order.Status == "closed" {
		return models.Order{}, fmt.Errorf("this order with %d is already closed", id)
	}

	// Шаг 1: Загружаем все позиции заказа
	var orderItems []models.OrderItem
	queryItems := `SELECT menu_item_id, quantity, price FROM order_items WHERE order_id = $1`
	rows, errItem := r.db.Query(queryItems, id)
	if errItem != nil {
		return models.Order{}, errItem
	}
	defer rows.Close()
	var item models.OrderItem
	for rows.Next() {
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
			return models.Order{}, err
		}
		orderItems = append(orderItems, item)
	}

	if err := rows.Err(); err != nil {
		return models.Order{}, err
	}
	// Загружаем все ингридиенты которые есть в позициях заказа
	var ingredients []models.MenuItemIngredient
	for _, item := range orderItems {
		queryIngredients := `SELECT ingredient_id, quantity  FROM menu_item_ingredients WHERE menu_item_id = $1`
		rows2, errIngredient := r.db.Query(queryIngredients, item.ProductID)
		if errIngredient != nil {
			return models.Order{}, errIngredient
		}
		defer rows2.Close()
		var ingredient models.MenuItemIngredient
		for rows2.Next() {
			if err := rows2.Scan(&ingredient.IngredientID, &ingredient.Quantity); err != nil {
				return models.Order{}, err
			}
			ingredients = append(ingredients, ingredient)
		}

		if err := rows2.Err(); err != nil {
			return models.Order{}, err
		}
	}

	// tm := db.NewTransactionManager(r.db)

	// errTransact := tm.WithTransaction(func(tx *db.TransactionManager) error {
	// Обновление инвентаря
	querySubsctruct := `UPDATE inventory SET quantity = quantity - $1 WHERE id = $2`
	for _, ingredient := range ingredients {
		// Используем Exec, передавая элементы по одному
		_, err := r.db.Exec(querySubsctruct, ingredient.Quantity, ingredient.IngredientID)
		if err != nil {
			return models.Order{}, fmt.Errorf("failed to update inventory: %v", err)
		}
	}

	queryClosing := `UPDATE orders SET status = $2, updated_at = NOW() WHERE id = $1`
	result, err := r.db.Exec(queryClosing, id, "closed")
	if err != nil {
		return models.Order{}, fmt.Errorf("error while closing order: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to get number of affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return models.Order{}, fmt.Errorf("order with ID %d not found", id)
	}

	// return  nil
	// })
	// if errTransact != nil {
	// 	return models.Order{}, errTransact
	// }
	order, errLoad = r.LoadOrder(id)
	if errLoad != nil {
		return models.Order{}, errLoad
	}

	return order, nil
}

func (r OrderRepository) GetOrderedItemsCount(start, end time.Time) (models.Order, error) {
	var orderedItems models.Order
	return orderedItems, nil
}
