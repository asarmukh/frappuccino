package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type CustomerRepository struct {
	db *sql.DB
}

type CustomerRepositoryInterface interface {
	FindByName(name string) (int, error)
	AddCustomer(name, phone string, preferences map[string]interface{}) (int, error)
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) FindByName(name string) (int, error) {
	var customerID int
	err := r.db.QueryRow("SELECT id FROM customers WHERE name = $1", name).Scan(&customerID)
	if err == sql.ErrNoRows {
		return 0, nil // Если клиента не нашли, возвращаем 0 и nil
	} else if err != nil {
		return 0, fmt.Errorf("error finding customer: %w", err)
	}
	return customerID, nil
}

func (r *CustomerRepository) AddCustomer(name, phone string, preferences map[string]interface{}) (int, error) {
	preferencesJSON, err := json.Marshal(preferences)
	if err != nil {
		return 0, fmt.Errorf("error serializing preferences: %w", err)
	}

	var customerID int
	err = r.db.QueryRow(
		"INSERT INTO customers (name, phone, preferences) VALUES ($1, $2, $3) RETURNING id",
		name, phone, preferencesJSON,
	).Scan(&customerID)
	if err != nil {
		return 0, fmt.Errorf("error creating customer: %w", err)
	}
	return customerID, nil
}
