package models

import (
	"time"
)

type Order struct {
	ID                  int               `json:"order_id"`
	CustomerName        string            `json:"customer_name"`
	Status              string            `json:"status"`
	TotalAmount         float64           `json:"total_amount,omitempty"`
	SpecialInstructions map[string]string `json:"special_instructions,omitempty"`
	Items               []OrderItem       `json:"items"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

type OrderItem struct {
	ProductID int     `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	Price     float64
}

type BulkOrderRequest struct {
	Orders []Order `json:"orders"`
}

type BulkOrderResponse struct {
	ProcessedOrders []struct {
		ID           int     `json:"order_id"`
		CustomerName string  `json:"customer_name"`
		Status       string  `json:"status"`
		TotalAmount  float64 `json:"total,omitempty"`  // Может отсутствовать, если заказ отклонён
		Reason       string  `json:"reason,omitempty"` // Причина отклонения
	} `json:"processed_orders"`

	Summary struct {
		TotalOrders      int     `json:"total_orders"`
		Accepted         int     `json:"accepted"`
		Rejected         int     `json:"rejected"`
		TotalRevenue     float64 `json:"total_revenue"`
		InventoryUpdates []struct {
			IngredientID int     `json:"ingredient_id"`
			Name         string  `json:"name"`
			QuantityUsed float64 `json:"quantity_used"`
			Remaining    float64 `json:"remaining"`
		} `json:"inventory_updates"`
	} `json:"summary"`
}
