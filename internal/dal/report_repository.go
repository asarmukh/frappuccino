package dal

import (
	"database/sql"
	"fmt"
	"frappuccino/models"
	"strings"
	"time"

	"github.com/lib/pq"
)

type ReportRepositoryInterface interface {
	TotalSales() (float64, error)
	GetPopularItems() ([]models.MenuItem, error)
	GetOrderedItemsByDay(month, year string) ([]map[string]int, error)
	GetOrderedItemsByMonth(year string) ([]map[string]int, error)
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

func (r ReportRepository) GetOrderedItemsByDay(month, year string) ([]map[string]int, error) {
	if month == "" {
		month = time.Now().Month().String() // Get the current month name
	}

	if year == "" {
		year = fmt.Sprintf("%d", time.Now().Year()) // Get the current year
	}

	// Convert month to title case to ensure it matches PostgreSQL's output format
	month = strings.Title(month)

	query := `
        SELECT EXTRACT(DAY FROM o.created_at) AS day, SUM(oi.quantity) AS quantity
        FROM order_items oi
        JOIN orders o ON oi.order_id = o.id
        WHERE TO_CHAR(o.created_at, 'Month') = $1
        AND EXTRACT(YEAR FROM o.created_at) = $2
        GROUP BY EXTRACT(DAY FROM o.created_at)
        ORDER BY day;
    `

	rows, err := r.db.Query(query, month, year)
	if err != nil {
		return nil, fmt.Errorf("failed to query ordered items by day: %w", err)
	}
	defer rows.Close()

	var orderedItems []map[string]int
	for rows.Next() {
		var day int
		var quantity int
		if err := rows.Scan(&day, &quantity); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		orderedItems = append(orderedItems, map[string]int{fmt.Sprintf("%d", day): quantity})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return orderedItems, nil
}

func (r ReportRepository) GetOrderedItemsByMonth(year string) ([]map[string]int, error) {
	if year == "" {
		year = fmt.Sprintf("%d", time.Now().Year())
	}

	query := `
		SELECT TO_CHAR(o.created_at, 'Month') AS month, SUM(oi.quantity) AS quantity
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		WHERE EXTRACT(YEAR FROM o.created_at) = $1
		GROUP BY TO_CHAR(o.created_at, 'Month')
		ORDER BY EXTRACT(MONTH FROM o.created_at);
	`

	rows, err := r.db.Query(query, year)
	if err != nil {
		return nil, fmt.Errorf("failed to query ordered items by month: %w", err)
	}
	defer rows.Close()

	var orderedItems []map[string]int
	for rows.Next() {
		var month string
		var quantity int
		if err := rows.Scan(&month, &quantity); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		orderedItems = append(orderedItems, map[string]int{month: quantity})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return orderedItems, nil
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
