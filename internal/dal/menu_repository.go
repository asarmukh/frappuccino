package dal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"frappuccino/models"
)

type MenuRepositoryInterface interface {
	AddMenuItem(menuItem models.MenuItem) error
	LoadMenuItems() ([]models.MenuItem, error)
	SaveMenuItems(menuItems []models.MenuItem) error
}

type MenuRepositoryJSON struct {
	filePath string
}

func NewMenuRepositoryJSON(filePath string) MenuRepositoryJSON {
	return MenuRepositoryJSON{filePath: filePath}
}

func (m MenuRepositoryJSON) AddMenuItem(menuItem models.MenuItem) error {
	menuItems, err := m.LoadMenuItems()
	if err != nil {
		return err
	}

	for _, item := range menuItems {
		if item.ID == menuItem.ID {
			return fmt.Errorf("menu item with ID %s already exists", menuItem.ID)
		}
	}

	menuItems = append(menuItems, menuItem)

	return m.SaveMenuItems(menuItems)
}

func (m MenuRepositoryJSON) LoadMenuItems() ([]models.MenuItem, error) {
	filePath := filepath.Join(m.filePath, "menu_items.json")
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.MenuItem{}, nil
		}
		return nil, fmt.Errorf("could not open menu file: %v", err)
	}
	defer file.Close()

	var menuItems []models.MenuItem
	if err := json.NewDecoder(file).Decode(&menuItems); err != nil && menuItems != nil {
		return nil, fmt.Errorf("could not decode menu: %v", err)
	}

	return menuItems, nil
}

func (r MenuRepositoryJSON) SaveMenuItems(menuItems []models.MenuItem) error {
	filePath := filepath.Join(r.filePath, "menu_items.json")
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("could not open or create inventory file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(menuItems); err != nil {
		return fmt.Errorf("could not encode inventory to file: %v", err)
	}

	return nil
}
