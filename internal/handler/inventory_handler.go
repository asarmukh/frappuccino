package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"frappuccino/internal/service"
	"frappuccino/models"
	"frappuccino/utils"
)

type InventoryHandlerInterface interface {
	HandleCreateInventory(w http.ResponseWriter, r *http.Request)
	HandleGetAllInventory(w http.ResponseWriter, r *http.Request)
	HandleGetInventoryById(w http.ResponseWriter, r *http.Request, id string)
	HandleDeleteInventoryItem(w http.ResponseWriter, r *http.Request, inventoryItemID string)
	HandleUpdateInventoryItem(w http.ResponseWriter, r *http.Request, inventoryItemID string)
}

type InventoryHandler struct {
	inventoryService service.InventoryService
}

func NewInventoryHandler(_inventoryService service.InventoryService) InventoryHandler {
	return InventoryHandler{inventoryService: _inventoryService}
}

func (h InventoryHandler) HandleCreateInventory(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request to create inventory")

	var newInventory models.InventoryItem
	if err := json.NewDecoder(r.Body).Decode(&newInventory); err != nil {
		slog.Warn("Invalid JSON format", "error", err)
		utils.ErrorInJSON(w, http.StatusBadRequest, fmt.Errorf("invalid request body JSON format: %v", err))
		return
	}

	inventory, err := h.inventoryService.CreateInventory(newInventory)
	if err != nil {
		slog.Error("Failed to create inventory", "error", err)
		utils.ErrorInJSON(w, http.StatusBadRequest, err)
		return
	} else {
		slog.Info("inventory created successfully", "IngredientID", inventory.IngredientID)
		utils.ResponseInJSON(w, 201, inventory)
	}
}

// func (h InventoryHandler) HandleGetAllInventory(w http.ResponseWriter, r *http.Request) {
// 	slog.Info("Received request to get all inventory")

// 	inventories, err := h.inventoryService.GetAllInventory()
// 	if err != nil {
// 		slog.Error("Failed to retrieve inventory", "error", err)
// 		utils.ErrorInJSON(w, http.StatusNotFound, err)
// 	}

// 	slog.Info("Successfully retrieved all inventory", "count", len(inventories))
// 	utils.ResponseInJSON(w, 200, inventories)
// }

// func (h InventoryHandler) HandleGetInventoryById(w http.ResponseWriter, r *http.Request, id string) {
// 	slog.Info("Received request to get inventory", "inventoryID", id)

// 	inventory, err := h.inventoryService.GetInventoryByID(id)
// 	if err != nil {
// 		slog.Warn("inventory not found", "inventoryID", id, "error", err)
// 		utils.ErrorInJSON(w, 404, err)
// 		return
// 	}

// 	slog.Info("Successfully retrieved inventory", "inventoryID", inventory.IngredientID)
// 	utils.ResponseInJSON(w, 200, inventory)
// }

// func (h InventoryHandler) HandleDeleteInventoryItem(w http.ResponseWriter, r *http.Request, inventoryItemID string) {
// 	slog.Info("Received request to delete inventory", "inevntoryID", inventoryItemID)

// 	err := h.inventoryService.DeleteInventoryItemByID(inventoryItemID)
// 	if err != nil {
// 		slog.Warn("Failed to delete inventory", "inventoryID", inventoryItemID, "error", err)
// 		utils.ErrorInJSON(w, http.StatusNotFound, err)
// 		return
// 	}

// 	slog.Info("inventory deleted successfully", "inventoryID", inventoryItemID)
// 	w.WriteHeader(http.StatusNoContent)
// }

// func (h InventoryHandler) HandleUpdateInventoryItem(w http.ResponseWriter, r *http.Request, inventoryItemID string) {
// 	slog.Info("Received request to update inventory", "inventoryID", inventoryItemID)

// 	var changedInventoryItem models.InventoryItem
// 	if err := json.NewDecoder(r.Body).Decode(&changedInventoryItem); err != nil {
// 		slog.Warn("Invalid JSON format", "error", err)
// 		utils.ErrorInJSON(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	item, err := h.inventoryService.UpdateInventoryItem(inventoryItemID, changedInventoryItem)
// 	if err != nil {
// 		slog.Warn("Failed to update inventory", "inventoryID", inventoryItemID, "error", err)
// 		utils.ErrorInJSON(w, http.StatusNotFound, err)
// 		return
// 	}
// 	slog.Info("inventory updated successfully", "inventoryID", item.IngredientID)
// 	utils.ResponseInJSON(w, 200, item)
// }
