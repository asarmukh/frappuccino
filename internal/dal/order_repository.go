package dal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"frappuccino/internal/database"
	"frappuccino/models"
	"frappuccino/utils"
)

type OrderRepositoryInterface interface {
	AddOrder(order models.Order) (models.Order, error)
	LoadOrders() ([]models.Order, error)
	LoadOrder(id int) (models.Order, error)
	DeleteOrderByID(id int) error
	UpdateOrder(id int) (models.Order, error)
	CloseOrder(id int) (models.Order, []struct {
		IngredientID int     `json:"ingredient_id"`
		Name         string  `json:"name"`
		QuantityUsed float64 `json:"quantity_used"`
		Remaining    float64 `json:"remaining"`
	}, error)
	GetOrderedItemsCount(start, end time.Time) (map[string]int, error)
}

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return OrderRepository{db: db}
}

// Method for adding a new order to the database
func (r OrderRepository) AddOrder(order models.Order) (models.Order, error) {
	query := `INSERT INTO orders (name, total_amount, special_instructions)
			VALUES ($1, $2, $3)
			RETURNING id, name, status, total_amount, special_instructions, created_at, updated_at`

	errTransact := database.WithTransaction(r.db, func(tx *sql.Tx) error {
		specialInstructionsByte, err := json.Marshal(order.SpecialInstructions)
		if err != nil {
			return err
		}
		var specialInstructionsData []byte
		if err := tx.QueryRow(
			query,
			order.CustomerName,
			order.TotalAmount,
			specialInstructionsByte,
		).Scan(&order.ID, &order.CustomerName, &order.Status, &order.TotalAmount, &specialInstructionsData, &order.CreatedAt, &order.UpdatedAt); err != nil {
			log.Printf("Error inserting order: %v", err)
			return err
		}
		// Декодируем JSONB в map[string]string
		if err := json.Unmarshal(specialInstructionsData, &order.SpecialInstructions); err != nil {
			return err
		}

		queryPrice := `SELECT price FROM menu_items WHERE id = $1`
		for i := range order.Items {
			err := tx.QueryRow(queryPrice, order.Items[i].ProductID).Scan(&order.Items[i].Price)
			if err != nil {
				return err
			}

			query := `INSERT INTO order_items (order_id, menu_item_id, quantity, price)
				  VALUES ($1, $2, $3, $4)
				  ON CONFLICT (order_id, menu_item_id) DO NOTHING`
			_, err = tx.Exec(query, order.ID, order.Items[i].ProductID, order.Items[i].Quantity, order.Items[i].Price)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if errTransact != nil {
		return models.Order{}, errTransact
	}

	queryItems := `SELECT menu_item_id, quantity, price FROM order_items WHERE order_id = $1`
	rows, err := r.db.Query(queryItems, order.ID)
	if err != nil {
		return models.Order{}, fmt.Errorf("error getting list of order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
			return models.Order{}, fmt.Errorf("error scanning items: %w", err)
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
		return nil, fmt.Errorf("Query execution error: %v", err)
	}
	defer rows.Close() // Закрываем `rows` только ОДИН раз

	for rows.Next() {
		var order models.Order
		var specialInstructionsStr string

		// Сканируем данные заказа
		if err := rows.Scan(&order.ID, &order.CustomerName, &order.Status, &order.TotalAmount, &specialInstructionsStr, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("line scan error: %v", err)
		}

		// Декодируем JSON
		if err := json.Unmarshal([]byte(specialInstructionsStr), &order.SpecialInstructions); err != nil {
			return nil, fmt.Errorf("decoding error JSON: %v", err)
		}

		// Загружаем товары для этого заказа
		queryItems := `SELECT menu_item_id, quantity, price FROM order_items WHERE order_id = $1`
		itemRows, err := r.db.Query(queryItems, order.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting list of order items: %w", err)
		}
		defer itemRows.Close() // Отдельный `defer` для товаров

		var items []models.OrderItem
		for itemRows.Next() {
			var item models.OrderItem
			if err := itemRows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
				return nil, fmt.Errorf("error scanning items: %w", err)
			}
			items = append(items, item)
		}
		order.Items = items

		orders = append(orders, order)
	}

	// Проверяем ошибки во внешнем rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
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
		return models.Order{}, fmt.Errorf("error getting list of order items: %w", err)
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
	errTransact := database.WithTransaction(r.db, func(tx *sql.Tx) error {
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
	})
	if errTransact != nil {
		return errTransact
	}

	return nil
}

// Обновление заказа
func (r OrderRepository) UpdateOrder(id int, changeOrder models.Order) (models.Order, error) {
	order, errLoad := r.LoadOrder(id)
	if errLoad != nil {
		return order, errLoad
	}
	if order.Status == "closed" {
		return models.Order{}, fmt.Errorf("this order with %d is closed", id)
	}

	var orderUpdated models.Order
	var specialInstructionsJSON []byte

	specialInstructionsBytes, err := json.Marshal(changeOrder.SpecialInstructions)
	if err != nil {
		return models.Order{}, fmt.Errorf("JSON marshaling error: %w", err)
	}

	queryUpdate := `
        UPDATE orders 
        SET name = $2, status = $3, total_amount = $4, special_instructions = $5::jsonb, updated_at = NOW() 
        WHERE id = $1 
        RETURNING *`

	errTransact := database.WithTransaction(r.db, func(tx *sql.Tx) error {
		err = tx.QueryRow(queryUpdate, id, changeOrder.CustomerName, "updated", changeOrder.TotalAmount, specialInstructionsBytes).
			Scan(&orderUpdated.ID, &orderUpdated.CustomerName, &orderUpdated.Status, &orderUpdated.TotalAmount, &specialInstructionsJSON, &orderUpdated.CreatedAt, &orderUpdated.UpdatedAt)
		if err != nil {
			return fmt.Errorf("request execution error: %w", err)
		}
		// Декодирование JSON обратно в карту
		err = json.Unmarshal(specialInstructionsJSON, &orderUpdated.SpecialInstructions)
		if err != nil {
			return fmt.Errorf("JSON unmarshaling error: %w", err)
		}

		// Обновление позиций заказа
		queryPrice := `SELECT price FROM menu_items WHERE id = $1`
		queryUpdateItems := `UPDATE order_items SET quantity = $3, price = $4 WHERE order_id = $1 AND menu_item_id = $2`
		for _, item := range changeOrder.Items {
			err := tx.QueryRow(queryPrice, item.ProductID).Scan(&item.Price)
			if err != nil {
				return fmt.Errorf("error getting a price: %w", err)
			}

			_, err = tx.Exec(queryUpdateItems, orderUpdated.ID, item.ProductID, item.Quantity, item.Price)
			if err != nil {
				return fmt.Errorf("order Item Update Error: %w", err)
			}
		}
		return nil
	})
	if errTransact != nil {
		return models.Order{}, errTransact
	}

	// Получение актуальных позиций заказа
	queryItems := `SELECT menu_item_id, quantity, price FROM order_items WHERE order_id = $1`
	rows, err := r.db.Query(queryItems, orderUpdated.ID)
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

	return orderUpdated, nil
}

func (r OrderRepository) CloseOrder(id int) (models.Order, []struct {
	IngredientID int     `json:"ingredient_id"`
	Name         string  `json:"name"`
	QuantityUsed float64 `json:"quantity_used"`
	Remaining    float64 `json:"remaining"`
}, error,
) {
	order, errLoad := r.LoadOrder(id)
	if errLoad != nil {
		return order, nil, errLoad
	}
	if order.Status == "closed" {
		return models.Order{}, nil, fmt.Errorf("this order with %d is already closed", id)
	}

	// Шаг 1: Загружаем все позиции заказа
	var orderItems []models.OrderItem
	queryItems := `SELECT menu_item_id, quantity, price FROM order_items WHERE order_id = $1`
	rows, errItem := r.db.Query(queryItems, id)
	if errItem != nil {
		return models.Order{}, nil, errItem
	}
	defer rows.Close()
	var item models.OrderItem
	for rows.Next() {
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
			return models.Order{}, nil, err
		}
		orderItems = append(orderItems, item)
	}

	if err := rows.Err(); err != nil {
		return models.Order{}, nil, err
	}

	// Загружаем все ингридиенты которые есть в позициях заказа
	var ingredients []models.MenuItemIngredient
	for _, item := range orderItems {
		queryIngredients := `SELECT ingredient_id, quantity FROM menu_item_ingredients WHERE menu_item_id = $1`
		rows2, errIngredient := r.db.Query(queryIngredients, item.ProductID)
		if errIngredient != nil {
			return models.Order{}, nil, errIngredient
		}
		defer rows2.Close()
		var ingredient models.MenuItemIngredient
		for rows2.Next() {
			if err := rows2.Scan(&ingredient.IngredientID, &ingredient.Quantity); err != nil {
				return models.Order{}, nil, err
			}
			ingredient.Quantity *= item.Quantity // Умножаем на количество предметов в заказе
			ingredients = append(ingredients, ingredient)
		}

		if err := rows2.Err(); err != nil {
			return models.Order{}, nil, err
		}
	}

	// Суммируем количество каждого ингредиента по ID
	ingredientQuantities := make(map[int]float64)
	for _, ing := range ingredients {
		ingredientQuantities[ing.IngredientID] += ing.Quantity
	}

	// Проверяем хватает ли в инвентаре ингридиентов для закрытия текущего заказа
	// и собираем информацию о текущих количествах
	queryCheck := `SELECT id, ingredient_name, quantity FROM inventory WHERE id = $1`

	type inventoryItem struct {
		ID       int
		Name     string
		Quantity float64
	}

	inventoryItems := make(map[int]inventoryItem)

	for ingredientID, requiredQuantity := range ingredientQuantities {
		var item inventoryItem
		err := r.db.QueryRow(queryCheck, ingredientID).Scan(&item.ID, &item.Name, &item.Quantity)
		if err != nil {
			return models.Order{}, nil, fmt.Errorf("failed to check inventory: %v", err)
		}
		if item.Quantity < requiredQuantity {
			return models.Order{}, nil, fmt.Errorf("insufficient inventory for ingredient ID %d (available: %f, required: %f)",
				ingredientID, item.Quantity, requiredQuantity)
		}
		inventoryItems[ingredientID] = item
	}

	// Создаем слайс для отслеживания обновлений инвентаря
	var inventoryUpdates []struct {
		IngredientID int     `json:"ingredient_id"`
		Name         string  `json:"name"`
		QuantityUsed float64 `json:"quantity_used"`
		Remaining    float64 `json:"remaining"`
	}

	errTransact := database.WithTransaction(r.db, func(tx *sql.Tx) error {
		// Обновление инвентаря
		querySubsctruct := `UPDATE inventory SET quantity = quantity - $1 WHERE id = $2 RETURNING quantity`
		for ingredientID, requiredQuantity := range ingredientQuantities {
			// Используем QueryRow вместо Exec, чтобы получить оставшееся количество
			var reorderThreshold float64
			var remaining float64
			err := tx.QueryRow(querySubsctruct, requiredQuantity, ingredientID).Scan(&remaining)
			if err != nil {
				return fmt.Errorf("failed to update inventory: %v", err)
			}

			// порог перезаказа:
			reorderThresholdQuery := `SELECT quantity, reorder_threshold FROM inventory WHERE id = $1`
			err = tx.QueryRow(reorderThresholdQuery, ingredientID).Scan(&remaining, &reorderThreshold)
			if err != nil {
				return fmt.Errorf("failed to check inventory: %v", err)
			}

			if remaining <= reorderThreshold {
				slog.Warn("⚠️ Warning!")
				slog.Warn("⚠️ Ingredient %d is below reorder threshold (%f left)\n", ingredientID, remaining)
				// Тут можно добавить логику создания заказа поставщику
			}

			// Добавляем информацию об обновлении в слайс
			inventoryUpdates = append(inventoryUpdates, struct {
				IngredientID int     `json:"ingredient_id"`
				Name         string  `json:"name"`
				QuantityUsed float64 `json:"quantity_used"`
				Remaining    float64 `json:"remaining"`
			}{
				IngredientID: ingredientID,
				Name:         inventoryItems[ingredientID].Name,
				QuantityUsed: requiredQuantity,
				Remaining:    remaining,
			})
		}

		queryClosing := `UPDATE orders SET status = $2, updated_at = NOW() WHERE id = $1`
		result, err := tx.Exec(queryClosing, id, "closed")
		if err != nil {
			return fmt.Errorf("error while closing order: %v", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get number of affected rows: %v", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("order with ID %d not found", id)
		}

		return nil
	})

	if errTransact != nil {
		return models.Order{}, nil, errTransact
	}

	order, errLoad = r.LoadOrder(id)
	if errLoad != nil {
		return models.Order{}, nil, errLoad
	}

	return order, inventoryUpdates, nil
}

func (r OrderRepository) GetOrderedItemsCount(start, end time.Time) (map[string]int, error) {
	query := `
		SELECT mi.name, SUM(oi.quantity) 
		FROM order_items oi
		JOIN menu_items mi ON oi.menu_item_id = mi.id
		JOIN orders o ON oi.order_id = o.id
		WHERE o.created_at BETWEEN $1 AND $2
		GROUP BY mi.name;
	`

	rows, err := r.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query ordered items count: %w", err)
	}
	defer rows.Close()

	orderedItems := make(map[string]int)
	for rows.Next() {
		var itemName string
		var quantity int
		if err := rows.Scan(&itemName, &quantity); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		orderedItems[itemName] = quantity
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return orderedItems, nil
}
