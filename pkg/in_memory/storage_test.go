package inmemory

import (
	"sync"
	"testing"
)

func TestStorage_Basic(t *testing.T) {
	s := NewStorage()

	t.Run("set_and_get", func(t *testing.T) {
		s.Set("key1", "value1")
		val, exists := s.Get("key1")
		if !exists || val != "value1" {
			t.Errorf("expected value1, got %s (exists: %v)", val, exists)
		}
	})

	t.Run("get_nonexistent", func(t *testing.T) {
		_, exists := s.Get("missing")
		if exists {
			t.Error("expected exists to be false for missing key")
		}
	})
}

func TestStorage_IncrAndGet(t *testing.T) {
	s := NewStorage()

	t.Run("sequential_inc", func(t *testing.T) {
		if val := s.IncrAndGet(); val != 1 {
			t.Errorf("expected 1, got %d", val)
		}
		if val := s.IncrAndGet(); val != 2 {
			t.Errorf("expected 2, got %d", val)
		}
	})

	t.Run("concurrent_inc", func(t *testing.T) {
		s = NewStorage() // reset
		const workers = 100
		const incsPerWorker = 100
		var wg sync.WaitGroup

		wg.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < incsPerWorker; j++ {
					s.IncrAndGet()
				}
			}()
		}
		wg.Wait()

		expected := uint64(workers * incsPerWorker)
		if val := s.IncrAndGet(); val != expected+1 {
			t.Errorf("expected %d, got %d", expected+1, val)
		}
	})
}

func TestStorage_Concurrency(t *testing.T) {
	s := NewStorage()
	const iterations = 1000
	var wg sync.WaitGroup

	wg.Add(2)

	// пишем параллельно
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			s.Set("key", "value")
		}
	}()

	// читаем параллельно
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			s.Get("key")
		}
	}()

	wg.Wait()
}