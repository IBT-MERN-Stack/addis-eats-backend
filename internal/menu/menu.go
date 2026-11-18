package menu

import "errors"

var ErrNotFound = errors.New("dish not found")

// Allowed sort fields for safety against invalid parameters
var SortFields = []string{"name", "-name", "price", "-price"}

type Dish struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PriceSantim int    `json:"price_santim"`
	IsFasting   bool   `json:"is_fasting"`
	CostSantim  int    `json:"-"`
}

// Query carries filter and sort specifications down to the store.
// Fasting uses a *bool pointer to distinguish between "not set" (nil) and "false" (&false).
type Query struct {
	Fasting *bool
	Sort    string
}

type DishReader interface {
	All() ([]Dish, error)
	ByID(id string) (Dish, error)
	Filter(q Query) ([]Dish, error)
}