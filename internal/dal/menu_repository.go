package dal

import (
	"database/sql"
	"frappuccino/models"
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
	var newMenuItem models.MenuItem

	query := `INSERT INTO menu_items
			(name, description, price, categories)
			VALUES ($1, $2, $3, $4)
			RETURNING id, name, description, price, categories, created_at, updated_at`
	err := r.db.QueryRow(
		query,
		menuItem.Name,
		menuItem.Description,
		menuItem.Price,
		menuItem.Categories).Scan(&newMenuItem.ID, &newMenuItem.Name, &newMenuItem.Description, &newMenuItem.Price, &newMenuItem.Categories, &newMenuItem.CreatedAt, &newMenuItem.UpdatedAt)
	if err != nil {
		return models.MenuItem{}, err
	}

	return newMenuItem, nil
}
