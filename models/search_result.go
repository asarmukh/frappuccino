package models

type SearchResult struct {
	MenuItems    []MenuItem `json:"menu_items"`
	Orders       []Order    `json:"orders"`
	TotalMatches int        `json:"total_matches"`
}
