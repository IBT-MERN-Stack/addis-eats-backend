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
	ByID(id string) (Order, error)
	UpdateStatus(id string, newStatus Status) (Order, error)
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

// newID generates cryptographically random hex IDs (`o-a1b2`).
// SECURITY NOTE: Avoid sequential integer IDs (1, 2, 3). Sequential IDs create
// Insecure Direct Object Reference (IDOR) vulnerabilities where attackers iterate through order IDs.
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

func (s *MemoryStore) ByID(id string) (Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	o, ok := s.orders[id]
	if !ok {
		return Order{}, ErrNotFound
	}
	return *o, nil
}

// UpdateStatus validates transitions and mutates order status atomically under a write-lock.
func (s *MemoryStore) UpdateStatus(id string, newStatus Status) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	o, ok := s.orders[id]
	if !ok {
		return Order{}, ErrNotFound
	}

	// Verify workflow transition rules
	if !o.Status.CanTransitionTo(newStatus) {
		return Order{}, fmt.Errorf("cannot transition from %s to %s: %w", o.Status, newStatus, ErrIllegalTransition)
	}

	o.Status = newStatus
	return *o, nil
}