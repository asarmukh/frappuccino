package dal

import (
	"database/sql"
	"frappuccino/models"
)

type InventoryRepositoryInterface interface {
	AddInventory(inventory models.InventoryItem) (models.InventoryItem, error)
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

// func (r InventoryRepositoryPostgres) SaveInventory(inventories []models.InventoryItem) error {
// 	filePath := filepath.Join(r.filePath, "inventory.json")
// 	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
// 	if err != nil {
// 		return fmt.Errorf("could not open or create inventory file: %v", err)
// 	}
// 	defer file.Close()

// 	encoder := json.NewEncoder(file)
// 	encoder.SetIndent("", "  ")
// 	if err := encoder.Encode(inventories); err != nil {
// 		return fmt.Errorf("could not encode inventory to file: %v", err)
// 	}

// 	return nil
// }

// func (r InventoryRepositoryPostgres) LoadInventory() ([]models.InventoryItem, error) {
// 	filePath := filepath.Join(r.filePath, "inventory.json")

// 	if _, err := os.Stat(filePath); os.IsNotExist(err) {
// 		return nil, fmt.Errorf("inventory file does not exist: %v", err)
// 	}

// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not open inventory file: %v", err)
// 	}
// 	defer file.Close()

// 	var inventories []models.InventoryItem
// 	if err := json.NewDecoder(file).Decode(&inventories); err != nil {
// 		return nil, err
// 	}

// 	return inventories, nil
// }
