package dal

import (
	"database/sql"
	"fmt"
	"frappuccino/models"
	"strings"
)

type InventoryRepositoryInterface interface {
	AddInventory(inventory models.InventoryItem) (models.InventoryItem, error)
	LoadInventory() ([]models.InventoryItem, error)
	SaveInventory(inventories []models.InventoryItem) error
}

type InventoryRepositoryPostgres struct {
	db *sql.DB
}

func NewInventoryPostgresRepository(_db *sql.DB) InventoryRepositoryPostgres {
	return InventoryRepositoryPostgres{db: _db}
}

func (r InventoryRepositoryPostgres) AddInventory(inventory models.InventoryItem) (models.InventoryItem, error) {
	var newInventory models.InventoryItem
	query := `INSERT INTO inventory
			(ingredient_name, quantity, unit, reorder_threshold)
			VALUES ($1, $2, $3, $4)
			RETURNING id, ingredient_name, quantity, unit, reorder_threshold`
	err := r.db.QueryRow(
		query,
		inventory.Name,
		inventory.Quantity,
		inventory.Unit,
		inventory.ReorderThreshold).Scan(&newInventory.IngredientID, &newInventory.Name, &newInventory.Quantity, &newInventory.Unit, &newInventory.ReorderThreshold)
	if err != nil {
		return models.InventoryItem{}, err
	}

	return newInventory, nil
}

func (r InventoryRepositoryPostgres) SaveInventory(inventories []models.InventoryItem) error {
	if len(inventories) == 0 {
		return nil
	}

	query := "INSERT INTO inventory (ingredient_name, quantity, unit, reorder_threshold) VALUES "
	placeholders := []string{}
	var values []interface{}

	for i, inv := range inventories {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		values = append(values, inv.Name, inv.Quantity, inv.Unit, inv.ReorderThreshold)
	}

	query += strings.Join(placeholders, ", ")

	_, err := r.db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении в базу данных: %v", err)
	}

	return nil
}

func (r InventoryRepositoryPostgres) LoadInventory() ([]models.InventoryItem, error) {
	var inventories []models.InventoryItem

	query := `SELECT id, ingredient_name, quantity, unit, reorder_threshold FROM inventory`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var inventory models.InventoryItem
		if err := rows.Scan(&inventory.IngredientID, &inventory.Name, &inventory.Quantity, &inventory.Unit, &inventory.ReorderThreshold); err != nil {
			return nil, err
		}
		inventories = append(inventories, inventory)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return inventories, nil
}
