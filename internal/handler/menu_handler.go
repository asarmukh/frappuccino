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

type MenuHandlerInterface interface {
	HandleCreateMenuItem(w http.ResponseWriter, r *http.Request)
	HandleGetAllMenuItems(w http.ResponseWriter, r *http.Request)
	HandleGetMenuItemById(w http.ResponseWriter, r *http.Request, menuID string)
	HandleDeleteMenuItemById(w http.ResponseWriter, r *http.Request, menuID string)
	HandleUpdateMenu(w http.ResponseWriter, r *http.Request, menuID string)
}

type MenuHandler struct {
	menuService service.MenuService
}

func NewMenuHandler(_menuService service.MenuService) MenuHandler {
	return MenuHandler{menuService: _menuService}
}

func (m MenuHandler) HandleCreateMenuItem(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request to add a new menu item")

	var NewMenuItem models.MenuItem
	if err := json.NewDecoder(r.Body).Decode(&NewMenuItem); err != nil {
		slog.Warn("Invalid JSON format", "error", err)
		utils.ErrorInJSON(w, http.StatusBadRequest, fmt.Errorf("invalid JSON format: %v", err))
		return
	}

	if NewMenuItem.ID != 0 { // 0 по умолчанию для int, чтобы в сервисы не заходить
		slog.Warn("invalid request body")
		utils.ErrorInJSON(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	if err := utils.ValidateMenuItem(NewMenuItem); err != nil {
		slog.Warn("Menu item validation failed", "error", err)
		utils.ErrorInJSON(w, http.StatusBadRequest, err)
		return
	}

	menu, err := m.menuService.CreateMenuItem(NewMenuItem)
	if err != nil {
		slog.Error("Failed to add menu item", "error", err)
		utils.ErrorInJSON(w, 400, err)
		return
	}

	slog.Info("Menu item added successfully", "menuID", menu.ID)
	utils.ResponseInJSON(w, 201, menu)
}

// func (m MenuHandler) HandleGetAllMenuItems(w http.ResponseWriter, r *http.Request) {
// 	slog.Info("Received request to get all menu items")

// 	items, err := m.menuService.GetAllMenuItems()
// 	if err != nil {
// 		slog.Error("Failed to retrieve menu items", "error", err)
// 		utils.ErrorInJSON(w, http.StatusInternalServerError, err)
// 		return
// 	}

// 	if len(items) == 0 {
// 		utils.ResponseInJSON(w, 200, []models.MenuItem{})
// 		return
// 	}

// 	slog.Info("Successfully retrieved all menu items", "count", len(items))
// 	utils.ResponseInJSON(w, 200, items)
// }

// func (m MenuHandler) HandleGetMenuItemById(w http.ResponseWriter, r *http.Request, menuID string) {
// 	slog.Info("Received request to get menu item", "menuID", menuID)

// 	if err := utils.ValidateID(menuID); err != nil {
// 		slog.Warn("Invalid menu ID", "menuID", menuID, "error", err)
// 		utils.ErrorInJSON(w, http.StatusBadRequest, fmt.Errorf("invalid menu ID: %v", err))
// 		return
// 	}

// 	item, err := m.menuService.GetMenuItemByID(menuID)
// 	if err != nil {
// 		slog.Warn("Menu item not found", "menuID", menuID, "error", err)
// 		utils.ErrorInJSON(w, http.StatusNotFound, err)
// 		return
// 	}

// 	slog.Info("Successfully retrieved menu item", "menuID", item.ID)
// 	utils.ResponseInJSON(w, 200, item)
// }

// func (m MenuHandler) HandleDeleteMenuItemById(w http.ResponseWriter, r *http.Request, menuID string) {
// 	slog.Info("Received request to delete menu item", "menuID", menuID)

// 	if err := utils.ValidateID(menuID); err != nil {
// 		slog.Warn("Invalid menu ID", "menuID", menuID, "error", err)
// 		utils.ErrorInJSON(w, http.StatusBadRequest, fmt.Errorf("invalid menu ID: %v", err))
// 		return
// 	}

// 	err := m.menuService.DeleteMenuItemByID(menuID)
// 	if err != nil {
// 		slog.Warn("Failed to delete menu item", "menuID", menuID, "error", err)
// 		utils.ErrorInJSON(w, http.StatusNotFound, err)
// 		return
// 	}

// 	slog.Info("Menu item deleted successfully", "menuID", menuID)
// 	w.WriteHeader(http.StatusNoContent)
// }

// func (m MenuHandler) HandleUpdateMenu(w http.ResponseWriter, r *http.Request, menuID string) {
// 	slog.Info("Received request to update menu item", "menuID", menuID)

// 	if err := utils.ValidateID(menuID); err != nil {
// 		slog.Warn("Invalid menu ID", "menuID", menuID, "error", err)
// 		utils.ErrorInJSON(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	var changeMenu models.MenuItem
// 	if err := json.NewDecoder(r.Body).Decode(&changeMenu); err != nil {
// 		slog.Warn("Invalid JSON format", "error", err)
// 		utils.ErrorInJSON(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	if err := utils.ValidateMenuItem(changeMenu); err != nil {
// 		slog.Warn("Menu item validation failed", "error", err)
// 		utils.ErrorInJSON(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	menu, err := m.menuService.UpdateMenu(menuID, changeMenu)
// 	if err != nil {
// 		slog.Warn("Failed to update menu item", "menuID", menuID, "error", err)
// 		utils.ErrorInJSON(w, http.StatusNotFound, err)
// 		return
// 	}

// 	slog.Info("Menu item updated successfully", "menuID", menu.ID)
// 	utils.ResponseInJSON(w, 200, menu)
// }
