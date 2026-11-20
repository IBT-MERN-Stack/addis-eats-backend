package menu

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// MemoryStore provides a thread-safe in-memory implementation of DishReader.
type MemoryStore struct {
	mu     sync.RWMutex // RWMutex allows concurrent reads while protecting writes
	dishes []Dish
}

// NewMemoryStore initializes the memory store with baseline data.
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
	dishes, _, err := s.Page(Query{Sort: "name"}, 1, 100)
	return dishes, err
}

func (s *MemoryStore) ByID(id string) (Dish, error) {
	s.mu.RLock() // Acquire read-lock for safe concurrent map/slice reading
	defer s.mu.RUnlock()

	for _, d := range s.dishes {
		if d.ID == id {
			return d, nil
		}
	}
	return Dish{}, ErrNotFound
}

// Page handles filtering, sorting, and offset windowing inside the store [11].
// TEACHING NOTE: Never fetch the entire dataset into HTTP handlers and slice in memory.
// Filtering and pagination belong inside the storage layer so it can be pushed down to SQL databases later [11].
func (s *MemoryStore) Page(q Query, page, limit int) ([]Dish, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 1. Apply Filtering
	var filtered []Dish
	for _, d := range s.dishes {
		if q.Fasting != nil && d.IsFasting != *q.Fasting {
			continue
		}
		filtered = append(filtered, d)
	}
	total := len(filtered)

	// 2. Apply Sorting
	sort.Slice(filtered, func(i, j int) bool {
		switch q.Sort {
		case "-name":
			return strings.Compare(filtered[i].Name, filtered[j].Name) > 0
		case "price":
			return filtered[i].PriceSantim < filtered[j].PriceSantim
		case "-price":
			return filtered[i].PriceSantim > filtered[j].PriceSantim
		case "name":
			fallthrough
		default:
			return strings.Compare(filtered[i].Name, filtered[j].Name) < 0
		}
	})

	// 3. Perform Bounded Slice Calculation (Offset Pagination) [12]
	offset := (page - 1) * limit
	if offset >= total {
		// Return empty non-nil slice so JSON encodes as `[]` instead of `null`
		return []Dish{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	// Defensive copy: return a newly allocated slice to prevent caller mutations
	result := make([]Dish, end-offset)
	copy(result, filtered[offset:end])
	return result, total, nil
}

// ETag computes a cryptographic hash representing the state of the queried collection [4].
// TEACHING NOTE: HTTP ETags allow clients to issue conditional requests (If-None-Match).
// If the collection state hasn't changed, the server can answer 304 Not Modified without sending any body [4, 5].
func (s *MemoryStore) ETag(q Query) string {
	dishes, total, _ := s.Page(q, 1, 1000)

	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d-%d", total, len(dishes))))
	for _, d := range dishes {
		h.Write([]byte(fmt.Sprintf("%s:%d;", d.ID, d.PriceSantim)))
	}

	return fmt.Sprintf(`W/"%x"`, h.Sum(nil)[:8]) // Formatted as weak ETag (W/"...")
}