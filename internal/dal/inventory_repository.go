package dal

import (
	"database/sql"
	"errors"
	"fmt"
	"frappuccino/internal/database"
	"frappuccino/models"
)

type InventoryRepositoryInterface interface {
	AddInventory(inventory models.InventoryItem) (models.InventoryItem, error)
	LoadInventory() ([]models.InventoryItem, error)
	GetInventoryItemByID(id int) (models.InventoryItem, error)
	DeleteInventoryItemByID(id int) error
	UpdateInventoryItem(inventoryItemID int, changedInventoryItem models.InventoryItem) (models.InventoryItem, error)
}

type InventoryRepositoryPostgres struct {
	db *sql.DB
}

func NewInventoryRepository(_db *sql.DB) InventoryRepositoryPostgres {
	return InventoryRepositoryPostgres{db: _db}
}

func (r InventoryRepositoryPostgres) AddInventory(inventory models.InventoryItem) (models.InventoryItem, error) {
	var newInventory models.InventoryItem

	query := `INSERT INTO inventory
	  (ingredient_name, quantity, unit, reorder_threshold)
	  VALUES ($1, $2, $3, $4)
	  RETURNING id, ingredient_name, quantity, unit, reorder_threshold`
	errTransact := database.WithTransaction(r.db, func(tx *sql.Tx) error {
		err := tx.QueryRow(
			query,
			inventory.Name,
			inventory.Quantity,
			inventory.Unit,
			inventory.ReorderThreshold,
		).Scan(&newInventory.IngredientID, &newInventory.Name, &newInventory.Quantity, &newInventory.Unit, &newInventory.ReorderThreshold)
		if err != nil {
			return fmt.Errorf("ошибка при выполнении запроса: %v", err)
		}
		return nil
	})
	if errTransact != nil {
		return models.InventoryItem{}, errTransact
	}

	return newInventory, nil
}

func (r InventoryRepositoryPostgres) LoadInventory() ([]models.InventoryItem, error) {
	var inventories []models.InventoryItem

	query := `SELECT id, ingredient_name, quantity, unit, reorder_threshold FROM inventory`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var inventory models.InventoryItem
		if err := rows.Scan(&inventory.IngredientID, &inventory.Name, &inventory.Quantity, &inventory.Unit, &inventory.ReorderThreshold); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки: %v", err)
		}
		inventories = append(inventories, inventory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации строк: %v", err)
	}

	return inventories, nil
}

func (r InventoryRepositoryPostgres) GetInventoryItemByID(id int) (models.InventoryItem, error) {
	var inventory models.InventoryItem

	query := `SELECT id, ingredient_name, quantity, unit, reorder_threshold FROM inventory WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&inventory.IngredientID,
		&inventory.Name,
		&inventory.Quantity,
		&inventory.Unit,
		&inventory.ReorderThreshold,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.InventoryItem{}, fmt.Errorf("inventory item с ID %d не найден", id)
		}
		return models.InventoryItem{}, fmt.Errorf("ошибка при получении элемента: %v", err)
	}

	return inventory, nil
}

func (r InventoryRepositoryPostgres) DeleteInventoryItemByID(id int) error {
	query := `DELETE FROM inventory WHERE id = $1`
	errTransact := database.WithTransaction(r.db, func(tx *sql.Tx) error {
		result, err := tx.Exec(query, id)
		if err != nil {
			return fmt.Errorf("ошибка при удалении элемента: %v", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("не удалось получить количество затронутых строк: %v", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("inventory item с ID %d не найден", id)
		}
		return nil
	})
	if errTransact != nil {
		return errTransact
	}

	return nil
}

func (r InventoryRepositoryPostgres) UpdateInventoryItem(inventoryItemID int, changedInventoryItem models.InventoryItem) (models.InventoryItem, error) {
	var existingItem models.InventoryItem
	query := `SELECT id, ingredient_name, quantity, unit, reorder_threshold FROM inventory WHERE id = $1`
	err := r.db.QueryRow(query, inventoryItemID).Scan(
		&existingItem.IngredientID,
		&existingItem.Name,
		&existingItem.Quantity,
		&existingItem.Unit,
		&existingItem.ReorderThreshold,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.InventoryItem{}, fmt.Errorf("inventory item с ID %d не найден", inventoryItemID)
		}
		return models.InventoryItem{}, fmt.Errorf("ошибка при получении элемента: %v", err)
	}

	if changedInventoryItem.IngredientID != 0 && changedInventoryItem.IngredientID != existingItem.IngredientID {
		return models.InventoryItem{}, errors.New("нельзя изменить ID")
	}

	errTransact := database.WithTransaction(r.db, func(tx *sql.Tx) error {
		updateQuery := `UPDATE inventory SET ingredient_name = $1, quantity = $2, unit = $3, reorder_threshold = $4, updated_at = NOW() WHERE id = $5 RETURNING id, ingredient_name, quantity, unit, reorder_threshold, updated_at`
		err = tx.QueryRow(
			updateQuery,
			changedInventoryItem.Name,
			changedInventoryItem.Quantity,
			changedInventoryItem.Unit,
			changedInventoryItem.ReorderThreshold,
			inventoryItemID,
		).Scan(
			&existingItem.IngredientID,
			&existingItem.Name,
			&existingItem.Quantity,
			&existingItem.Unit,
			&existingItem.ReorderThreshold,
			&existingItem.UpdatedAt,
		)

		if err != nil {
			return fmt.Errorf("ошибка при обновлении элемента: %v", err)
		}
		return nil
	})
	if errTransact != nil {
		return models.InventoryItem{}, errTransact
	}

	return existingItem, nil
}
