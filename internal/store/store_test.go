package store

import (
	"sync"
	"testing"
)

func TestIDsStartAtOneAndIncrease(t *testing.T) {
	s := New()
	for want := 1; want <= 3; want++ {
		if got := s.Add(Transaction{Description: "x"}).ID; got != want {
			t.Fatalf("id = %d; want %d", got, want)
		}
	}
}

func TestResetClearsAndRestartsIDs(t *testing.T) {
	s := New()
	s.Add(Transaction{})
	s.Add(Transaction{})
	s.Reset()
	if n := len(s.All()); n != 0 {
		t.Fatalf("%d transactions after reset; want 0", n)
	}
	if id := s.Add(Transaction{}).ID; id != 1 {
		t.Errorf("first id after reset = %d; want 1", id)
	}
}

func TestAllReturnsACopy(t *testing.T) {
	s := New()
	s.Add(Transaction{Description: "coffee"})
	s.All()[0].Description = "changed"
	if got := s.All()[0].Description; got != "coffee" {
		t.Errorf("store was changed through All(): %q", got)
	}
}

func TestConcurrentAddsGetUniqueIDs(t *testing.T) {
	s := New()
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Add(Transaction{})
		}()
	}
	wg.Wait()
	seen := map[int]bool{}
	for _, tx := range s.All() {
		if seen[tx.ID] {
			t.Fatalf("id %d given twice", tx.ID)
		}
		seen[tx.ID] = true
	}
	if len(seen) != 100 {
		t.Errorf("%d transactions; want 100", len(seen))
	}
}
