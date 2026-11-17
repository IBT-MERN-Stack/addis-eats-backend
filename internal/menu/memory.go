package menu

import "sync"

type MemoryStore struct {
	mu     sync.RWMutex
	dishes []Dish
}

// NewMemoryStore initializes the 8 standard Addis Eats dishes.
func NewMemoryStore() DishReader {
	return &MemoryStore{
		dishes: []Dish{
			{ID: "d1", Name: "Doro Wat", PriceSantim: 18500, IsFasting: false, CostSantim: 12000},
			{ID: "d2", Name: "Shiro", PriceSantim: 12000, IsFasting: true, CostSantim: 6000},
			{ID: "d3", Name: "Tibs", PriceSantim: 24000, IsFasting: false, CostSantim: 15000},
			{ID: "d4", Name: "Injera Basket", PriceSantim: 4500, IsFasting: true, CostSantim: 2000},
			{ID: "d5", Name: "Misir Wot", PriceSantim: 11000, IsFasting: true, CostSantim: 5500},
			{ID: "d6", Name: "Gomen", PriceSantim: 9500, IsFasting: true, CostSantim: 4000},
			{ID: "d7", Name: "Kitfo", PriceSantim: 26000, IsFasting: false, CostSantim: 17000},
			{ID: "d8", Name: "Bayeinetu", PriceSantim: 16000, IsFasting: true, CostSantim: 9000},
		},
	}
}

func (s *MemoryStore) All() ([]Dish, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Initialized with make to prevent JSON null serialization on empty slices
	dst := make([]Dish, 0, len(s.dishes))
	dst = append(dst, s.dishes...)
	return dst, nil
}

func (s *MemoryStore) ByID(id string) (Dish, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, d := range s.dishes {
		if d.ID == id {
			return d, nil
		}
	}
	return Dish{}, ErrNotFound
}