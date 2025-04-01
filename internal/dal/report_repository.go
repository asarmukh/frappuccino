package dal

import (
	"database/sql"
	"frappuccino/models"
	"strings"

	"github.com/lib/pq"
)

type ReportRepositoryInterface interface {
	TotalSales() (float64, error)
	GetPopularItems() ([]models.MenuItem, error)
	SearchMenu(q string, minPrice int, maxPrice int) ([]models.MenuItem, error)
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

func (r ReportRepository) SearchMenu(q string, minPrice int, maxPrice int) ([]models.MenuItem, error) {
	// Разбиваем поисковый запрос на слова и подготавливаем их для SQL
	words := strings.Fields(q)
	for i, word := range words {
		words[i] = "%" + word + "%"
	}

	// SQL-запрос без динамических аргументов
	query := `
		SELECT id, name, description, price
		FROM menu_items
		WHERE (name ILIKE ANY($1) OR description ILIKE ANY($1))
		AND price BETWEEN $2 AND $3
	`

	// Выполняем запрос
	rows, err := r.db.Query(query, pq.Array(words), minPrice, maxPrice)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Обрабатываем результаты
	var items []models.MenuItem
	for rows.Next() {
		var item models.MenuItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
