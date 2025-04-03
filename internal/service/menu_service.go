package service

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"frappuccino/internal/dal"
	"frappuccino/models"
	"frappuccino/utils"
)

type MenuServiceInterface interface {
	CreateMenuItem(menuItem models.MenuItem) (models.MenuItem, error)
	GetAllMenuItems() ([]models.MenuItem, error)
	GetMenuItemByID(id int) (models.MenuItem, error)
	DeleteMenuItemByID(id int) error
	UpdateMenu(id int, changeMenu models.MenuItem) (models.MenuItem, error)
}

type MenuService struct {
	repository dal.MenuRepositoryInterface
}

func NewMenuService(_repository dal.MenuRepositoryInterface) MenuService {
	return MenuService{repository: _repository}
}

func (s MenuService) CreateMenuItem(menuItem models.MenuItem) (models.MenuItem, error) {
	if err := utils.ValidateMenuItem(menuItem); err != nil {
		return models.MenuItem{}, err
	}

	newMenuItem, err := s.repository.AddMenuItem(menuItem)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return models.MenuItem{}, errors.New("menu item with this name already exists")
		}
		return models.MenuItem{}, err
	}

	// }
	// 3)искать айди в бд продукта
	// if err := utils.ValidateID(menuItem.ID); err != nil {
	// 	return models.MenuItem{}, fmt.Errorf("invalid product ID: %v", err)
	// }
	// 4)искать айди в бд ингридиентов инвентаря
	// if err := utils.ValidateMenuItem(menuItem); err != nil {
	// 	return models.MenuItem{}, err
	// }

	log.Printf("menu item added: %d", newMenuItem.ID)
	return newMenuItem, nil
}

func (m MenuService) GetAllMenuItems() ([]models.MenuItem, error) {
	items, err := m.repository.LoadMenuItems()
	if err != nil {
		log.Printf("could not load menu items: %v", err)
		return nil, fmt.Errorf("could not load menu items: %v", err)
	}
	return items, nil
}

func (m MenuService) GetMenuItemByID(id int) (models.MenuItem, error) {
	// if err := utils.ValidateID(id); err != nil {
	// 	return models.MenuItem{}, fmt.Errorf("invalid menu ID: %v", err)
	// }

	return m.repository.GetMenuItemByID(id)
}

func (m MenuService) DeleteMenuItemByID(id int) error {
	// if err := utils.ValidateID(id); err != nil {
	// 	return fmt.Errorf("invalid menu ID: %v", err)
	// }

	return m.repository.DeleteMenuItemByID(id)
}

func (m MenuService) UpdateMenu(id int, changeMenu models.MenuItem) (models.MenuItem, error) {
	return m.repository.UpdateMenu(id, changeMenu)
}
