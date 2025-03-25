package models

import "time"

type Order struct {
	ID                  int                    `json:"order_id"`
	CustomerName        string                 `json:"customer_name"`
	Status              string                 `json:"status"`
	TotalAmount         float64                `json:"total_amount"`
	SpecialInstructions map[string]interface{} `json:"special_instructions,omitempty"`
	IsCompleted         bool                   `json:"is_completed"`
	Items               []OrderItem            `json:"items"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

type OrderItem struct {
	ProductID int     `json:"product_id"`
	Quantity  float64 `json:"quantity"`
}
