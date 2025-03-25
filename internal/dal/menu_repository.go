package dal

import (
	"database/sql"
	"errors"
	"frappuccino/models"
	"strings"
)

type MenuRepositoryInterface interface {
	AddMenuItem(menuItem models.MenuItem) (models.MenuItem, error)
	// LoadMenuItems() ([]models.MenuItem, error)
	// SaveMenuItems(menuItems []models.MenuItem) error
}

type MenuRepository struct {
	db *sql.DB
}

func NewMenuRepository(_db *sql.DB) MenuRepository {
	return MenuRepository{db: _db}
}

func (r MenuRepository) AddMenuItem(menuItem models.MenuItem) (models.MenuItem, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return models.MenuItem{}, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	categories := "{" + strings.Join(menuItem.Categories, ",") + "}"

	query := `INSERT INTO menu_items (name, description, price, categories) 
			  VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	err = tx.QueryRow(query, menuItem.Name, menuItem.Description, menuItem.Price, categories).
		Scan(&menuItem.ID, &menuItem.CreatedAt, &menuItem.UpdatedAt)
	if err != nil {
		return models.MenuItem{}, err
	}

	// Проверяем инвентарь на наличие которые были отправлены пользователем
	query2 := `SELECT EXISTS(SELECT 1 FROM inventory WHERE id = $1)`
	for _, ingredient := range menuItem.Ingredients {
		var exists bool
		err = tx.QueryRow(query2, ingredient.IngredientID).Scan(&exists)
		if err != nil {
			return models.MenuItem{}, err
		}
		if !exists {
			return models.MenuItem{}, errors.New("ingredient not found in inventory")
		}
	}

	// Вставка в menu_item_ingredients
	for _, ingredient := range menuItem.Ingredients {
		query := `INSERT INTO menu_item_ingredients (menu_item_id, ingredient_id, quantity)
				  VALUES ($1, $2, $3)`
		_, err = tx.Exec(query, menuItem.ID, ingredient.IngredientID, ingredient.Quantity)
		if err != nil {
			return models.MenuItem{}, err
		}
	}

	return menuItem, nil
}
