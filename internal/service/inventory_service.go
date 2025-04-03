package service

import (
	"errors"
	"frappuccino/internal/dal"
	"frappuccino/models"
	"frappuccino/utils"
	"log"
	"strings"
)

type InventoryServiceInterface interface {
	CreateInventory(Inventory models.InventoryItem) (models.InventoryItem, error)
	GetAllInventory() ([]models.InventoryItem, error)
	GetInventoryByID(id int) (models.InventoryItem, error)
	DeleteInventoryItemByID(id int) error
	UpdateInventoryItem(inventoryItemID int, changedInventoryItem models.InventoryItem) (models.InventoryItem, error)
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
	return s.repository.GetInventoryItemByID(id)
}

func (h InventoryService) DeleteInventoryItemByID(id int) error {
	return h.repository.DeleteInventoryItemByID(id)
}

func (h InventoryService) UpdateInventoryItem(inventoryItemID int, changedInventoryItem models.InventoryItem) (models.InventoryItem, error) {
	if changedInventoryItem.Quantity < 0 {
		return models.InventoryItem{}, errors.New("you can't pass a negative amount")
	}

	return h.repository.UpdateInventoryItem(inventoryItemID, changedInventoryItem)
}
