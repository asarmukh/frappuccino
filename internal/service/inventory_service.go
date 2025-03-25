package service

import (
	"errors"
	"fmt"
	"frappuccino/internal/dal"
	"frappuccino/models"
	"frappuccino/utils"
	"log"
	"strings"
)

type InventoryServiceInterface interface {
	CreateInventory(Inventory models.InventoryItem) (models.InventoryItem, error)
	GetAllInventory() ([]models.InventoryItem, error)
	GetInventoryByID(id string) (models.InventoryItem, error)
	DeleteInventoryItemByID(is string) error
	UpdateInventoryItem(inventoryItemID string, changedInventoryItem models.InventoryItem) (models.InventoryItem, error)
}

type InventoryService struct {
	repository dal.InventoryRepositoryInterface
}

func NewInventoryService(_repository dal.InventoryRepositoryInterface) InventoryService {
	return InventoryService{repository: _repository}
}

func (s InventoryService) CreateInventory(inventory models.InventoryItem) (models.InventoryItem, error) {
	if err := utils.IsValidName(inventory.Name); err != nil {
		return models.InventoryItem{}, err
	}

	if inventory.Name == "" || inventory.Quantity == 0 || inventory.Unit == "" {
		return models.InventoryItem{}, errors.New("invalid request body")
	}

	newInventory, err := s.repository.AddInventory(inventory)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return models.InventoryItem{}, errors.New("inventory item with this name already exists")
		}
		return models.InventoryItem{}, err
	}

	log.Printf("Inventory created: %d", newInventory.IngredientID)
	return newInventory, nil
}

func (s InventoryService) GetAllInventory() ([]models.InventoryItem, error) {
	inventrories, err := s.repository.LoadInventory()
	if err != nil {
		return nil, err
	}
	return inventrories, nil
}

func (s InventoryService) GetInventoryByID(id int) (models.InventoryItem, error) {
	inventory, err := s.repository.LoadInventory()
	if err != nil {
		return models.InventoryItem{}, errors.New("get all inventory")
	}

	if len(inventory) > 0 {
		for i := 0; i < len(inventory); i++ {
			if inventory[i].IngredientID == id {
				return inventory[i], nil
			}
		}
	} else {
		return models.InventoryItem{}, errors.New("inventory is empty")
	}
	return models.InventoryItem{}, errors.New("invalid product ID in inventory items")
}

func (h InventoryService) DeleteInventoryItemByID(id int) error {
	inventoryItems, err := h.repository.LoadInventory()
	if err != nil {
		return err
	}

	indexToDelete := -1
	for i, item := range inventoryItems {
		if item.IngredientID == id {
			indexToDelete = i
			break
		}
	}

	if indexToDelete == -1 {
		return fmt.Errorf("inventory item with ID %d not found", id)
	}

	inventoryItems = append(inventoryItems[:indexToDelete], inventoryItems[indexToDelete+1:]...)

	return h.repository.SaveInventory(inventoryItems)
}

// func (h InventoryService) UpdateInventoryItem(inventoryItemID string, changedInventoryItem models.InventoryItem) (models.InventoryItem, error) {
// 	if changedInventoryItem.Name == "" || changedInventoryItem.Quantity < 0 || changedInventoryItem.Unit == "" {
// 		return models.InventoryItem{}, errors.New("invalid request body")
// 	}

// 	inventoryItem, err := h.repository.LoadInventory()
// 	if err != nil {
// 		return models.InventoryItem{}, errors.New("invalid load inventory items")
// 	}

// 	var itemToUpdate *models.InventoryItem
// 	for i := 0; i < len(inventoryItem); i++ {
// 		if inventoryItem[i].IngredientID == inventoryItemID {
// 			if changedInventoryItem.IngredientID != inventoryItem[i].IngredientID {
// 				return models.InventoryItem{}, errors.New("cannot change ingredient ID")
// 			}
// 			inventoryItem[i].Name = changedInventoryItem.Name
// 			inventoryItem[i].Quantity = changedInventoryItem.Quantity
// 			inventoryItem[i].Unit = changedInventoryItem.Unit
// 			itemToUpdate = &inventoryItem[i]
// 			break
// 		}
// 	}
// 	if itemToUpdate == nil {
// 		return models.InventoryItem{}, errors.New("invalid ID in inventory items")
// 	}

// 	if err := h.repository.SaveInventory(inventoryItem); err != nil {
// 		return models.InventoryItem{}, errors.New("failed to save updated inventory")
// 	}

// 	return *itemToUpdate, nil
// }
