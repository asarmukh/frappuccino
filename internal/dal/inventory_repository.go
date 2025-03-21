package dal

import (
	"encoding/json"
	"fmt"
	"frappuccino/models"
	"os"
	"path/filepath"
)

type InventoryRepositoryInterface interface {
	AddInventory(inventory models.InventoryItem) error
	LoadInventory() ([]models.InventoryItem, error)
	SaveInventory(inventories []models.InventoryItem) error
}

type InventoryRepositoryJSON struct {
	filePath string
}

func NewInventoryRepositoryJSON(filepath string) InventoryRepositoryJSON {
	return InventoryRepositoryJSON{filePath: filepath}
}

func (r InventoryRepositoryJSON) AddInventory(inventory models.InventoryItem) error {
	inventories, err := r.LoadInventory()

	if err != nil && inventories != nil {
		return err
	}

	inventories = append(inventories, inventory)

	return r.SaveInventory(inventories)
}

func (r InventoryRepositoryJSON) SaveInventory(inventories []models.InventoryItem) error {
	filePath := filepath.Join(r.filePath, "inventory.json")
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("could not open or create inventory file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(inventories); err != nil {
		return fmt.Errorf("could not encode inventory to file: %v", err)
	}

	return nil
}

func (r InventoryRepositoryJSON) LoadInventory() ([]models.InventoryItem, error) {
	filePath := filepath.Join(r.filePath, "inventory.json")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("inventory file does not exist: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open inventory file: %v", err)
	}
	defer file.Close()

	var inventories []models.InventoryItem
	if err := json.NewDecoder(file).Decode(&inventories); err != nil {
		return nil, err
	}

	return inventories, nil
}
