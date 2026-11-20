package menu

import "errors"

// ErrNotFound is exported as a sentinel error so caller packages (like HTTP handlers)
// can inspect errors using `errors.Is(err, menu.ErrNotFound)` without needing raw string matching.
var ErrNotFound = errors.New("dish not found")

// SortFields acts as a strict whitelist for collection sorting parameters.
// TEACHING NOTE: Never build dynamic queries or sorts directly from raw user input.
// Validating against a whitelist prevents arbitrary parameter injection [10].
var SortFields = []string{"name", "-name", "price", "-price"}

// Dish represents a domain model item on the menu.
// STRUCT TAGS (-): Notice `CostSantim` is tagged as `json:"-"`. This ensures internal profit
// margins or wholesale costs are never leaked over the public JSON network boundary.
type Dish struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceSantim int   `json:"price_santim"` // Money is represented in integer santim to avoid floating-point errors
	IsFasting  bool   `json:"is_fasting"`
	CostSantim int    `json:"-"` // Excluded from public JSON serialization
}

// Query encapsulates filtering and sorting parameters passed down to the storage layer [11].
// TEACHING NOTE (*bool): Fasting uses a pointer `*bool` rather than a raw `bool`.
// This allows the code to distinguish between three states:
// 1. nil => Parameter was omitted (do not filter by fasting).
// 2. &true => Filter for fasting dishes only.
// 3. &false => Filter for non-fasting dishes only.
type Query struct {
	Fasting *bool
	Sort    string
}

// DishReader defines the storage contract for menu items.
// Handlers depend on this interface rather than a concrete struct, adhering to the
// Dependency Inversion Principle and enabling easy unit testing with mock stores.
type DishReader interface {
	All() ([]Dish, error)
	ByID(id string) (Dish, error)
	Page(q Query, page, limit int) ([]Dish, int, error)
	ETag(q Query) string
}