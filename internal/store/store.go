// Package store keeps transactions in memory. It is safe for concurrent use.
package store

import (
	"sync"

	"atad-project/internal/money"
)

type Transaction struct {
	ID          int          `json:"id"`
	Date        string       `json:"date"`
	Amount      money.Amount `json:"amount"`
	Description string       `json:"description"`
	Category    string       `json:"category"`
}

type Store struct {
	mu     sync.Mutex
	txs    []Transaction
	nextID int
}

func New() *Store {
	return &Store{nextID: 1}
}

// Add assigns the next id to t, stores it and returns the stored copy.
func (s *Store) Add(t Transaction) Transaction {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.ID = s.nextID
	s.nextID++
	s.txs = append(s.txs, t)
	return t
}

// All returns a copy of every transaction, in the order they were added.
func (s *Store) All() []Transaction {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Transaction, len(s.txs))
	copy(out, s.txs)
	return out
}

// Reset discards every transaction and starts the ids from 1 again.
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.txs = nil
	s.nextID = 1
}
