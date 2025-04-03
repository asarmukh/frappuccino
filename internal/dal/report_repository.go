package dal

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"frappuccino/models"

	"github.com/lib/pq"
)

type ReportRepositoryInterface interface {
	TotalSales() (float64, error)
	GetPopularItems() ([]models.MenuItem, error)
	GetOrderedItemsByDay(month string) ([]models.OrderItemReport, error)
	GetOrderedItemsByMonth(year int) ([]models.OrderItemReport, error)
	SearchMenu(q string, minPrice int, maxPrice int) (models.SearchResult, error)
	SearchOrders(q string, minPrice int, maxPrice int) (models.SearchResult, error)
}

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(_db *sql.DB) ReportRepository {
	return ReportRepository{db: _db}
}

func (r ReportRepository) TotalSales() (float64, error) {
	var orderClosed models.Order
	var totalSales float64
	queryTotalSales := `SELECT total_amount
	FROM orders
	WHERE status = 'closed'`

	rows, err := r.db.Query(queryTotalSales)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&orderClosed.TotalAmount); err != nil {
			return 0, err
		}
		totalSales += orderClosed.TotalAmount
	}

	if err := rows.Err(); err != nil {
		return 0, err
	}
	return totalSales, nil
}

func (r ReportRepository) GetPopularItems() ([]models.MenuItem, error) {
	query := `
	SELECT m.id, m.name, m.description, m.price, m.categories, m.created_at, m.updated_at 
	FROM menu_items m
	WHERE m.id IN (
		SELECT menu_item_id FROM order_items GROUP BY menu_item_id ORDER BY COUNT(menu_item_id) DESC
	)`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var popularItems []models.MenuItem
	for rows.Next() {
		var menu models.MenuItem
		if err := rows.Scan(
			&menu.ID,
			&menu.Name,
			&menu.Description,
			&menu.Price,
			pq.Array(&menu.Categories),
			&menu.CreatedAt,
			&menu.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Получение ингредиентов для текущего меню
		ingredientsQuery := `
		SELECT ingredient_id, quantity
		FROM menu_item_ingredients
		WHERE menu_item_id = $1`

		ingredientRows, err := r.db.Query(ingredientsQuery, menu.ID)
		if err != nil {
			return nil, err
		}
		defer ingredientRows.Close()

		// Присваиваем ингредиенты
		var ingredients []models.MenuItemIngredient
		for ingredientRows.Next() {
			var ingredient models.MenuItemIngredient
			if err := ingredientRows.Scan(&ingredient.IngredientID, &ingredient.Quantity); err != nil {
				return nil, err
			}
			ingredients = append(ingredients, ingredient)
		}

		// Присваиваем список ингредиентов в меню
		menu.Ingredients = ingredients
		popularItems = append(popularItems, menu)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return popularItems, nil
}

func (r *ReportRepository) GetOrderedItemsByDay(month string) ([]models.OrderItemReport, error) {
	if month == "" {
		return nil, fmt.Errorf("month is required when period is 'day'")
	}

	monthInt, err := time.Parse("January", month)
	if err != nil {
		return nil, fmt.Errorf("invalid month name: %w", err)
	}
	monthNumber := int(monthInt.Month())

	query := `
        SELECT EXTRACT(DAY FROM created_at) AS day, COUNT(*) 
        FROM orders 
        WHERE EXTRACT(MONTH FROM created_at) = $1        
        GROUP BY day
        ORDER BY day;
    `

	rows, err := r.db.Query(query, monthNumber)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer rows.Close()

	var result []models.OrderItemReport
	for rows.Next() {
		var day string
		var count int
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}

		result = append(result, models.OrderItemReport{
			Period: day,
			Count:  count,
		})
	}

	return result, nil
}

func (r *ReportRepository) GetOrderedItemsByMonth(year int) ([]models.OrderItemReport, error) {
	if year == 0 {
		return nil, fmt.Errorf("year is required when period is 'month'")
	}

	query := `
        SELECT TO_CHAR(created_at, 'Month') AS month, COUNT(*) 
        FROM orders 
        WHERE EXTRACT(YEAR FROM created_at) = $1
        GROUP BY month
        ORDER BY MIN(created_at);
    `

	rows, err := r.db.Query(query, year)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer rows.Close()

	var result []models.OrderItemReport
	for rows.Next() {
		var month string
		var count int
		if err := rows.Scan(&month, &count); err != nil {
			return nil, err
		}

		result = append(result, models.OrderItemReport{
			Period: strings.TrimSpace(month),
			Count:  count,
		})
	}

	return result, nil
}

func (r ReportRepository) SearchMenu(q string, minPrice int, maxPrice int) (models.SearchResult, error) {
	words := strings.Fields(q)
	for i, word := range words {
		words[i] = "%" + word + "%"
	}

	query := `
		SELECT id, name, description, price
		FROM menu_items
		WHERE (name ILIKE ANY($1) OR description ILIKE ANY($1))
		AND price BETWEEN $2 AND $3
	`

	rows, err := r.db.Query(query, pq.Array(words), minPrice, maxPrice)
	if err != nil {
		return models.SearchResult{}, err
	}
	defer rows.Close()

	var searchResult models.SearchResult

	// Создаем алиас для анонимной структуры
	type MenuItemType = struct {
		ID          int     `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
	}

	for rows.Next() {
		var item MenuItemType
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price); err != nil {
			return models.SearchResult{}, err
		}
		searchResult.MenuItems = append(searchResult.MenuItems, item)
		searchResult.TotalMatches++
	}

	return searchResult, nil
}

func (r ReportRepository) SearchOrders(q string, minPrice int, maxPrice int) (models.SearchResult, error) {
	words := strings.Fields(q)
	for i, word := range words {
		words[i] = "%" + word + "%"
	}

	queryOrders := `
    SELECT o.id, o.name AS customer_name, 
           array_agg(m.name) AS items, 
           o.total_amount AS total
    FROM orders o
    JOIN order_items oi ON o.id = oi.order_id
    JOIN menu_items m ON oi.menu_item_id = m.id
    WHERE (o.name ILIKE ANY($1) OR m.name ILIKE ANY($1))
    AND o.total_amount BETWEEN $2 AND $3
    GROUP BY o.id;
`

	rows, err := r.db.Query(queryOrders, pq.Array(words), minPrice, maxPrice)
	if err != nil {
		return models.SearchResult{}, err
	}
	defer rows.Close()

	var searchResult models.SearchResult

	// Создаем алиас для анонимной структуры
	type OrderType = struct {
		ID           int      `json:"id"`
		CustomerName string   `json:"customer_name"`
		Items        []string `json:"items"`
		Total        float64  `json:"total"`
	}

	for rows.Next() {
		var order OrderType
		if err := rows.Scan(&order.ID, &order.CustomerName, pq.Array(&order.Items), &order.Total); err != nil {
			return models.SearchResult{}, err
		}
		searchResult.Orders = append(searchResult.Orders, order)
		searchResult.TotalMatches++
	}

	return searchResult, nil
}
