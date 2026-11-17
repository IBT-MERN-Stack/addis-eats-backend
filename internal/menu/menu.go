package menu

import "errors"

// ErrNotFound is returned when a dish ID does not exist in the store.
var ErrNotFound = errors.New("dish not found")




// Dish represents a menu item with price stored strictly in integer santim.
type Dish struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PriceSantim int    `json:"price_santim"`
	IsFasting   bool   `json:"is_fasting"`
	CostSantim  int    `json:"-"` // Excluded from JSON output
}

// DishReader defines the read contract for dishes.
type DishReader interface {
	All() ([]Dish, error)
	ByID(id string) (Dish, error)
}





