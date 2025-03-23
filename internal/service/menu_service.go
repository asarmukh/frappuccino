package service

import (
	"frappuccino/internal/dal"
	"frappuccino/models"
)

type MenuServiceInterface interface {
	CreateMenuItem(menuItem models.MenuItem) (models.MenuItem, error)
	GetAllMenuItems() ([]models.MenuItem, error)
	GetMenuItemByID(id string) (models.MenuItem, error)
	DeleteMenuItemByID(id string) error
	UpdateMenu(id string, changeMenu models.MenuItem) (models.MenuItem, error)
}

type MenuService struct {
	repository       dal.MenuRepositoryInterface
	inventoryService InventoryService
}

func NewMenuService(_repository dal.MenuRepositoryInterface, _inventoryService InventoryService) MenuService {
	return MenuService{
		repository:       _repository,
		inventoryService: _inventoryService,
	}
}

// func (m MenuService) CreateMenuItem(menuItem models.MenuItem) (models.MenuItem, error) {
// 	words := strings.Fields(menuItem.Name)
// 	var newID string

// 	for i := 0; i < len(words); i++ {
// 		if i == len(words)-1 {
// 			newID = strings.ToLower(words[i]) // последний - ID
// 		}
// 	}

// 	menu, err := m.repository.LoadMenuItems()
// 	if err != nil {
// 		return models.MenuItem{}, err
// 	}

// 	for _, item := range menu {
// 		if item.Name == menuItem.Name || item.Description == menuItem.Description {
// 			return models.MenuItem{}, fmt.Errorf("invalid requests body: name or description exeist in menu")
// 		}
// 	}

// 	menuItem.ID = newID

// 	if err := utils.ValidateID(menuItem.ID); err != nil {
// 		return models.MenuItem{}, fmt.Errorf("invalid product ID: %v", err)
// 	}

// 	if err := utils.ValidateMenuItem(menuItem); err != nil {
// 		return models.MenuItem{}, err
// 	}

// 	if err := m.repository.AddMenuItem(menuItem); err != nil {
// 		return models.MenuItem{}, err
// 	}
// 	log.Printf("menu item added: %s", menuItem.ID)
// 	return menuItem, nil
// }

// func (m MenuService) GetAllMenuItems() ([]models.MenuItem, error) {
// 	items, err := m.repository.LoadMenuItems()
// 	if err != nil {
// 		log.Printf("could not load menu items: %v", err)
// 		return nil, fmt.Errorf("could not load menu items: %v", err)
// 	}
// 	return items, nil
// }

// func (m MenuService) GetMenuItemByID(id string) (models.MenuItem, error) {
// 	if err := utils.ValidateID(id); err != nil {
// 		return models.MenuItem{}, fmt.Errorf("invalid menu ID: %v", err)
// 	}

// 	menuItems, err := m.repository.LoadMenuItems()
// 	if err != nil {
// 		return models.MenuItem{}, fmt.Errorf("could not load menu items: %v", err)
// 	}

// 	for _, item := range menuItems {
// 		if item.ID == id {
// 			return item, nil
// 		}
// 	}

// 	return models.MenuItem{}, fmt.Errorf("menu item with ID %s not found", id)
// }

// func (m MenuService) DeleteMenuItemByID(id string) error {
// 	if err := utils.ValidateID(id); err != nil {
// 		return fmt.Errorf("invalid menu ID: %v", err)
// 	}

// 	menuItems, err := m.repository.LoadMenuItems()
// 	if err != nil {
// 		return fmt.Errorf("could not load menu items")
// 	}

// 	indexToDelete := -1
// 	for i, item := range menuItems {
// 		if item.ID == id {
// 			indexToDelete = i
// 			break
// 		}
// 	}

// 	if indexToDelete == -1 {
// 		return fmt.Errorf("menu item with ID %s not found", id)
// 	}

// 	menuItems = append(menuItems[:indexToDelete], menuItems[indexToDelete+1:]...)

// 	if err := m.repository.SaveMenuItems(menuItems); err != nil {
// 		return fmt.Errorf("could not save menu items")
// 	}

// 	return nil
// }

// func (m MenuService) UpdateMenu(id string, changeMenu models.MenuItem) (models.MenuItem, error) {
// 	if err := utils.ValidateID(id); err != nil {
// 		return models.MenuItem{}, err
// 	}

// 	if err := utils.ValidateMenuItem(changeMenu); err != nil {
// 		return models.MenuItem{}, err
// 	}

// 	inventory, err := m.inventoryService.GetAllInventory()
// 	if err != nil {
// 		return models.MenuItem{}, errors.New("unable to load inventory items")
// 	}

// 	inventoryMap := make(map[string]models.InventoryItem)
// 	for _, item := range inventory {
// 		inventoryMap[item.IngredientID] = item
// 	}

// 	for _, ingredient := range changeMenu.Ingredients {
// 		if _, exists := inventoryMap[ingredient.IngredientID]; !exists {
// 			return models.MenuItem{}, fmt.Errorf("ingredient with ID %s not found in inventory", ingredient.IngredientID)
// 		}
// 	}

// 	menu, err := m.repository.LoadMenuItems()
// 	if err != nil {
// 		return models.MenuItem{}, errors.New("unable to load menu items")
// 	}

// 	for i := 0; i < len(menu); i++ {
// 		if menu[i].ID == id {
// 			if changeMenu.ID != menu[i].ID {
// 				return models.MenuItem{}, errors.New("cannot change the ID of the menu item")
// 			}
// 			menu[i].Name = changeMenu.Name
// 			menu[i].Ingredients = changeMenu.Ingredients
// 			menu[i].Description = changeMenu.Description
// 			menu[i].Price = changeMenu.Price
// 			m.repository.SaveMenuItems(menu)
// 			return menu[i], nil
// 		}
// 	}
// 	return models.MenuItem{}, errors.New("menu item not found")
// }
