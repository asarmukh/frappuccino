package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"frappuccino/models"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"
)

var ErrValidation = errors.New("validation error")

type ErrorResponse struct {
	Error string `json:"error"`
}

func ErrorInJSON(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{Error: err.Error()}

	json.NewEncoder(w).Encode(response)
}

func ResponseInJSON(w http.ResponseWriter, statusCode int, object interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(object); err != nil {
		http.Error(w, `{"error": "Failed to encode JSON"}`, http.StatusInternalServerError)
	}
}

func ValidateMenuItem(menuItem models.MenuItem) error {
	if err := IsValidName(menuItem.Name); err != nil {
		return fmt.Errorf("%w: invalid name: %v", ErrValidation, err)
	}

	if err := ValidateDescription(menuItem.Description); err != nil {
		return fmt.Errorf("%w: invalid description: %v", ErrValidation, err)
	}

	if err := ValidatePrice(menuItem.Price); err != nil {
		return fmt.Errorf("%w: invalid price: %v", ErrValidation, err)
	}

	if err := ValidateIngredients(menuItem.Ingredients); err != nil {
		return fmt.Errorf("%w: invalid ingredients: %v", ErrValidation, err)
	}

	return nil
}

func ValidateDescription(description string) error {
	if description == "" {
		return fmt.Errorf("description cannot be empty")
	}

	length := utf8.RuneCountInString(description)
	if length < 10 || length > 500 {
		return fmt.Errorf("description length must be between 10 and 500 characters")
	}

	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	if htmlRegex.MatchString(description) {
		return fmt.Errorf("description cannot contain HTML tags")
	}

	return nil
}

func ValidatePrice(price float64) error {
	if price <= 0 {
		return fmt.Errorf("price must be greater than zero")
	}

	if price > 1000000 {
		return fmt.Errorf("price is too high")
	}

	return nil
}

func ValidateIngredients(ingredients []models.MenuItemIngredient) error {
	if len(ingredients) == 0 {
		return fmt.Errorf("ingredients list cannot be empty")
	}

	if len(ingredients) > 50 {
		return fmt.Errorf("too many ingredients (maximum 50)")
	}

	seenIngredients := make(map[int]bool)
	for _, ingredient := range ingredients {
		if err := ValidateIngredient(ingredient); err != nil {
			return err
		}

		if seenIngredients[ingredient.IngredientID] {
			return fmt.Errorf("duplicate ingredient ID: %d", ingredient.IngredientID)
		}
		seenIngredients[ingredient.IngredientID] = true
	}

	return nil
}

func ValidateIngredient(ingredient models.MenuItemIngredient) error {
	if ingredient.IngredientID <= 0 {
		return fmt.Errorf("invalid ingredient ID: %d", ingredient.IngredientID)
	}

	if ingredient.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}

	if ingredient.Quantity > 1000 {
		return fmt.Errorf("quantity is too high (maximum 1000)")
	}

	return nil
}

func ValidateID(id string) error {
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	validID := regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)
	if !validID.MatchString(id) {
		return fmt.Errorf("ID can only contain letters, numbers, and hyphens")
	}

	length := utf8.RuneCountInString(id)
	if length < 1 || length > 36 {
		return fmt.Errorf("ID length must be between 1 and 36 characters")
	}

	return nil
}

func IsValidName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	length := len(name)
	if length < 2 || length > 63 {
		return fmt.Errorf("name length must be between 2 and 63 characters")
	}

	validName := regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9\s'&()]+[a-zA-Z0-9]$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("name must start and end with letter or number and can contain only letters, numbers, spaces, hyphens, apostrophes, ampersands, and parentheses")
	}

	if strings.Contains(name, "  ") {
		return fmt.Errorf("name cannot contain consecutive spaces")
	}

	if strings.Contains(name, "--") {
		return fmt.Errorf("name cannot contain consecutive hyphens")
	}

	return nil
}

func ValidateSpecialInstructions(instructions map[string]string) error {
	allowedKeys := map[string]bool{
		"temperature": true,
		"notes":       true,
	}
	for key := range instructions {
		if !allowedKeys[key] {
			return fmt.Errorf("invalid key in special_instructions: %s", key)
		}
	}
	return nil
}

// Функция для преобразования special_instructions в map[string]string
func ConvertSpecialInstructions(data string) (map[string]string, error) {
	var result map[string]string
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling special_instructions: %v", err)
	}
	return result, nil
}

// func ValidateOrder(menu []models.MenuItem, getOrder models.Order) error {
// 	var idshki []string
// 	for i := 0; i < len(menu); i++ {
// 		for _, item := range getOrder.Items {
// 			if item.ProductID == menu[i].ID {
// 				idshki = append(idshki, item.ProductID)
// 			}
// 		}
// 	}
// 	var count int = len(idshki)
// 	var correctProducts []string
// 	// 2. Перебираем заказ и пробиваем на валидацию
// 	for _, item := range getOrder.Items {
// 		err := ValidateQuantity(float64(item.Quantity)) // преждевременная валидация на большие и отрицательные цифрры она не плоха
// 		if err != nil {
// 			return err
// 		}

// 		if len(idshki) == 0 {
// 			return fmt.Errorf("Invalid product ID: %s. Product not found in the menu.", item.ProductID)
// 		}

// 		for i := 0; i < len(idshki); i++ {
// 			if item.ProductID == idshki[i] {
// 				if len(getOrder.Items) == len(idshki) {
// 					count--
// 				}
// 				correctProducts = append(correctProducts, item.ProductID)
// 			}
// 		}
// 	}

// 	for _, item := range getOrder.Items {
// 		for _, product := range correctProducts {
// 			if count != 0 && item.ProductID != product {
// 				return fmt.Errorf("Invalid product ID %s. Product not found in the menu.", item.ProductID)
// 			}
// 		}
// 	}
// 	return nil
// }
