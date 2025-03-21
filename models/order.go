package models

import "time"

type Order struct {
	ID                  int                    `json:"order_id"`
	CustomerName        string                 `json:"customer_name"`
	CustomerPhone       string                 `json:"customer_phone"`
	CustomerPreferences map[string]interface{} `json:"customer_preferences"`
	Items               []OrderItem            `json:"items"`
	Status              string                 `json:"status"`
	TotalAmount         float64                `json:"total_amount"`
	SpecialInstructions string                 `json:"special_instructions,omitempty"`
	PaymentMethod       string                 `json:"payment_method"`
	IsCompleted         bool                   `json:"is_completed"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}
