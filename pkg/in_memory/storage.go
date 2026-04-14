package inmemory

import (
	"sync"
	"sync/atomic"
)

type Storage struct {
	mu      sync.RWMutex
	data    map[string]string
	counter uint64
}

func NewStorage() *Storage {
	return &Storage{data: make(map[string]string)}
}

func (s *Storage) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
}

func (s *Storage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.data[key]
	return val, ok
}

func (s *Storage) IncrAndGet() uint64 {
	return atomic.AddUint64(&s.counter, 1)
}
