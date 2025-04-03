package service

import (
	"errors"
	"log"
	"strings"

	"frappuccino/internal/dal"
	"frappuccino/models"
	"frappuccino/utils"
)

type InventoryServiceInterface interface {
	CreateInventory(Inventory models.InventoryItem) (models.InventoryItem, error)
	GetAllInventory() ([]models.InventoryItem, error)
	GetInventoryByID(id int) (models.InventoryItem, error)
	DeleteInventoryItemByID(id int) error
	UpdateInventoryItem(inventoryItemID int, changedInventoryItem models.InventoryItem) (models.InventoryItem, error)
	GetLeftovers(sortBy string, page, pageSize int) (map[string]interface{}, error)
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

func (h InventoryService) GetLeftovers(sortBy string, page, pageSize int) (map[string]interface{}, error) {
	if page < 1 {
		return nil, errors.New("page must be 1 or greater")
	}
	if pageSize < 1 {
		return nil, errors.New("pageSize must be 1 or greater")
	}

	items, totalCount, err := h.repository.GetLeftovers(sortBy, page, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	hasNextPage := page < totalPages

	response := map[string]interface{}{
		"currentPage": page,
		"hasNextPage": hasNextPage,
		"pageSize":    pageSize,
		"totalPages":  totalPages,
		"data":        items,
	}
	return response, nil
}
