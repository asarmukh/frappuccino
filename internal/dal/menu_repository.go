package dal

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"frappuccino/internal/database"
	"frappuccino/models"

	"github.com/lib/pq"
)

type MenuRepositoryInterface interface {
	AddMenuItem(menuItem models.MenuItem) (models.MenuItem, error)
	LoadMenuItems() ([]models.MenuItem, error)
	GetMenuItemByID(id int) (models.MenuItem, error)
	DeleteMenuItemByID(id int) error
	UpdateMenu(id int, changeMenu models.MenuItem) (models.MenuItem, error)
}

type MenuRepository struct {
	db *sql.DB
}

func NewMenuRepository(_db *sql.DB) MenuRepository {
	return MenuRepository{db: _db}
}

func (r MenuRepository) AddMenuItem(menuItem models.MenuItem) (models.MenuItem, error) {
	categories := "{" + strings.Join(menuItem.Categories, ",") + "}"

	errTransact := database.WithTransaction(r.db, func(tx *sql.Tx) error {
		query := `INSERT INTO menu_items (name, description, price, categories) 
			  VALUES ($1, $2, $3, $4) RETURNING id, created_at`
		err := tx.QueryRow(query, menuItem.Name, menuItem.Description, menuItem.Price, categories).
			Scan(&menuItem.ID, &menuItem.CreatedAt)
		if err != nil {
			return err
		}

		query2 := `SELECT EXISTS(SELECT 1 FROM inventory WHERE id = $1)`
		for _, ingredient := range menuItem.Ingredients {
			var exists bool
			err = tx.QueryRow(query2, ingredient.IngredientID).Scan(&exists)
			if err != nil {
				return err
			}
			if !exists {
				return errors.New("ingredient not found in inventory")
			}
		}

		for _, ingredient := range menuItem.Ingredients {
			query := `INSERT INTO menu_item_ingredients (menu_item_id, ingredient_id, quantity)
				  VALUES ($1, $2, $3)`
			_, err = tx.Exec(query, menuItem.ID, ingredient.IngredientID, ingredient.Quantity)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if errTransact != nil {
		return models.MenuItem{}, errTransact
	}

	return menuItem, nil
}

func (r MenuRepository) LoadMenuItems() ([]models.MenuItem, error) {
	var menuItems []models.MenuItem

	query := `SELECT id, name, description, price, categories, created_at
		FROM menu_items`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса для меню: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var menuItem models.MenuItem

		if err := rows.Scan(&menuItem.ID, &menuItem.Name, &menuItem.Description, &menuItem.Price,
			pq.Array(&menuItem.Categories), &menuItem.CreatedAt); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки меню: %v", err)
		}

		ingredientsQuery := `SELECT ingredient_id, quantity FROM menu_item_ingredients WHERE menu_item_id = $1`
		ingredientRows, err := r.db.Query(ingredientsQuery, menuItem.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при выполнении запроса для ингредиентов: %v", err)
		}
		defer ingredientRows.Close()

		var ingredients []models.MenuItemIngredient
		for ingredientRows.Next() {
			var ingredient models.MenuItemIngredient
			if err := ingredientRows.Scan(&ingredient.IngredientID, &ingredient.Quantity); err != nil {
				return nil, fmt.Errorf("ошибка при сканировании ингредиента: %v", err)
			}

			var inventoryQuantity float64
			inventoryQuery := `SELECT quantity FROM inventory WHERE id = $1`
			err := r.db.QueryRow(inventoryQuery, ingredient.IngredientID).Scan(&inventoryQuantity)
			if err != nil {
				return nil, fmt.Errorf("ошибка при проверке наличия ингредиента в инвентаре: %v", err)
			}

			if inventoryQuantity < ingredient.Quantity {
				return nil, fmt.Errorf("недостаточно ингредиентов %d в инвентаре", ingredient.IngredientID)
			}

			ingredients = append(ingredients, ingredient)
		}

		menuItem.Ingredients = ingredients
		menuItems = append(menuItems, menuItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации строк меню: %v", err)
	}

	return menuItems, nil
}

func (r MenuRepository) GetMenuItemByID(id int) (models.MenuItem, error) {
	var menuItem models.MenuItem

	query := `SELECT id, name, description, price, categories, created_at, updated_at
		FROM menu_items WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&menuItem.ID,
		&menuItem.Name,
		&menuItem.Description,
		&menuItem.Price,
		pq.Array(&menuItem.Categories),
		&menuItem.CreatedAt,
		&menuItem.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.MenuItem{}, fmt.Errorf("menu item с ID %d не найден", id)
		}
		return models.MenuItem{}, fmt.Errorf("ошибка при получении элемента: %v", err)
	}

	ingredientsQuery := `SELECT ingredient_id, quantity FROM menu_item_ingredients WHERE menu_item_id = $1`
	ingredientRows, err := r.db.Query(ingredientsQuery, menuItem.ID)
	if err != nil {
		return models.MenuItem{}, fmt.Errorf("ошибка при выполнении запроса для ингредиентов: %v", err)
	}
	defer ingredientRows.Close()

	var ingredients []models.MenuItemIngredient
	for ingredientRows.Next() {
		var ingredient models.MenuItemIngredient
		if err := ingredientRows.Scan(&ingredient.IngredientID, &ingredient.Quantity); err != nil {
			return models.MenuItem{}, fmt.Errorf("ошибка при сканировании ингредиента: %v", err)
		}

		ingredients = append(ingredients, ingredient)
	}

	menuItem.Ingredients = ingredients

	if err := ingredientRows.Err(); err != nil {
		return models.MenuItem{}, fmt.Errorf("ошибка при итерации ингредиентов: %v", err)
	}

	return menuItem, nil
}

func (r MenuRepository) ProductExists(productID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM menu_items WHERE id = $1)`
	var exists bool
	err := r.db.QueryRow(query, productID).Scan(&exists)
	return exists, err
}

func (r MenuRepository) GetProductPrice(productID int) (float64, error) {
	query := `SELECT price FROM menu_items WHERE id = $1`
	var price float64
	err := r.db.QueryRow(query, productID).Scan(&price)
	return price, err
}

func (r MenuRepository) DeleteMenuItemByID(id int) error {
	errTransact := database.WithTransaction(r.db, func(tx *sql.Tx) error {
		query := `DELETE FROM menu_items WHERE id = $1`
		res, err := tx.Exec(query, id)
		if err != nil {
			return fmt.Errorf("ошибка при удалении элемента: %v", err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("не удалось получить количество затронутых строк: %v", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("inventory item с ID %d не найден", id)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("ошибка при коммите транзакции: %v", err)
		}
		return nil
	})
	if errTransact != nil {
		return errTransact
	}

	return nil
}

func (r MenuRepository) UpdateMenu(id int, changeMenu models.MenuItem) (models.MenuItem, error) {
	var existingItem models.MenuItem

	query := `SELECT id, name, description, price, categories FROM menu_items WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&existingItem.ID,
		&existingItem.Name,
		&existingItem.Description,
		&existingItem.Price,
		pq.Array(&existingItem.Categories),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.MenuItem{}, fmt.Errorf("menu item с ID %d не найден", id)
		}
		return models.MenuItem{}, fmt.Errorf("ошибка при получении элемента: %v", err)
	}

	if changeMenu.ID != 0 && changeMenu.ID != existingItem.ID {
		return models.MenuItem{}, errors.New("нельзя изменить ID")
	}

	errTransact := database.WithTransaction(r.db, func(tx *sql.Tx) error {
		updateQuery := `UPDATE menu_items SET name = $1, description = $2, price = $3, categories = $4, updated_at = NOW() WHERE id = $5 RETURNING id, name, description, price, categories, created_at, updated_at`
		err = tx.QueryRow(
			updateQuery,
			changeMenu.Name,
			changeMenu.Description,
			changeMenu.Price,
			pq.Array(&changeMenu.Categories),
			id,
		).Scan(
			&existingItem.ID,
			&existingItem.Name,
			&existingItem.Description,
			&existingItem.Price,
			pq.Array(&existingItem.Categories),
			&existingItem.CreatedAt,
			&existingItem.UpdatedAt,
		)

		for _, ingredient := range changeMenu.Ingredients {
			queryUpdateIngredients := `UPDATE menu_item_ingredients 
		SET quantity = $3 
		WHERE menu_item_id = $1 AND ingredient_id = $2 
		RETURNING ingredient_id, quantity`

			err = tx.QueryRow(queryUpdateIngredients, id, ingredient.IngredientID, ingredient.Quantity).Scan(&ingredient.IngredientID, &ingredient.Quantity)
			if err != nil {
				return fmt.Errorf("ошибка при обновлении ингредиента: %v", err)
			}
			existingItem.Ingredients = append(existingItem.Ingredients, ingredient)
		}

		if err != nil {
			return fmt.Errorf("ошибка при обновлении элемента: %v", err)
		}
		return nil
	})
	if errTransact != nil {
		return models.MenuItem{}, errTransact
	}

	return existingItem, nil
}
