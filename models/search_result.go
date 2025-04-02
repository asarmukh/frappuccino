package models

type SearchResult struct {
	MenuItems []struct {
		ID          int     `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
	} `json:"menu_items"`

	Orders []struct {
		ID           int      `json:"id"`
		CustomerName string   `json:"customer_name"`
		Items        []string `json:"items"`
		Total        float64  `json:"total"`
	} `json:"orders"`

	TotalMatches int `json:"total_matches"`
}
