package models

import "time"

type MenuItem struct {
	ID          int                  `json:"product_id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Price       float64              `json:"price"`
	Categories  []string             `json:"categories,omitempty"`
	Available   bool                 `json:"available"`
	Ingredients []MenuItemIngredient `json:"ingredients"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type MenuItemIngredient struct {
	IngredientID int     `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
}
