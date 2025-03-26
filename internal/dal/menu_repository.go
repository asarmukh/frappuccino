package dal

import (
	"database/sql"
	"errors"
	"frappuccino/models"
	"log"
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
			log.Println("Transaction rolled back:", err)
		} else {
			err = tx.Commit()
			if err != nil {
				log.Println("Commit error:", err)
			}
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

	for _, ingredient := range menuItem.Ingredients {
		query := `INSERT INTO menu_item_ingredients (menu_item_id, ingredient_id, quantity)
				  VALUES ($1, $2, $3)`
		_, err = tx.Exec(query, menuItem.ID, ingredient.IngredientID, ingredient.Quantity)
		if err != nil {
			return models.MenuItem{}, err
		}
	}

	updateQuery := `UPDATE inventory SET quantity = quantity - $1 WHERE id = $2`
	for _, ingredient := range menuItem.Ingredients {
		_, err = tx.Exec(updateQuery, ingredient.Quantity, ingredient.IngredientID)
		if err != nil {
			return models.MenuItem{}, err
		}
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
