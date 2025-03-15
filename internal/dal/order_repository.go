package dal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"frappuccino/models"
)

type OrderRepositoryInterface interface {
	AddOrder(order models.Order) error
	LoadOrders() ([]models.Order, error)
	SaveOrders(orders []models.Order) error
}

type OrderRepositoryJSON struct {
	filePath string
}

func NewOrderRepositoryJSON(filePath string) OrderRepositoryJSON {
	return OrderRepositoryJSON{filePath: filePath}
}

func (r OrderRepositoryJSON) AddOrder(order models.Order) error {
	orders, err := r.LoadOrders()
	if err != nil && orders != nil {
		return err
	}

	orders = append(orders, order)

	return r.SaveOrders(orders)
}

func (r OrderRepositoryJSON) LoadOrders() ([]models.Order, error) {
	filePath := filepath.Join(r.filePath, "orders.json")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var orders []models.Order
	if err := json.NewDecoder(file).Decode(&orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r OrderRepositoryJSON) SaveOrders(orders []models.Order) error {
	filePath := filepath.Join(r.filePath, "orders.json")
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("could not open or create inventory file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(orders); err != nil {
		return fmt.Errorf("could not encode inventory to file: %v", err)
	}

	return nil
}
