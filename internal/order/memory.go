package order

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type Store interface {
	Create(o *Order) error
}

type MemoryStore struct {
	mu     sync.RWMutex
	orders map[string]*Order
}



func NewMemoryStore() Store {
	return &MemoryStore{
		orders: make(map[string]*Order),
	}
}


func newID() string {
	b := make([]byte, 2)
	_, _ = rand.Read(b)
	return fmt.Sprintf("o-%s", hex.EncodeToString(b))
}

func (s *MemoryStore) Create(o *Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	o.ID = newID()
	o.PlacedAt = time.Now().UTC()
	s.orders[o.ID] = o
	return nil
}