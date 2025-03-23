package models

import "time"

type InventoryItem struct {
	IngredientID     int       `json:"ingredient_id"`
	Name             string    `json:"name"`
	Quantity         float64   `json:"quantity"`
	Unit             string    `json:"unit"`
	ReorderThreshold *float64  `json:"reorder_threshold,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
}
