package modules

import (
	"sync"
	"testing"
)

func TestInterruptDataAddWebsite(t *testing.T) {
	id := &InterruptData{}

	site1 := []string{"200", "192.168.1.1", "Test Site 1"}
	site2 := []string{"200", "192.168.1.2", "Test Site 2"}

	id.AddWebsite(site1)
	id.AddWebsite(site2)

	websites := id.GetWebsites()

	if len(websites) != 2 {
		t.Errorf("Expected 2 websites, got %d", len(websites))
	}

	if websites[0][2] != "Test Site 1" {
		t.Errorf("Expected 'Test Site 1', got %s", websites[0][2])
	}

	if websites[1][2] != "Test Site 2" {
		t.Errorf("Expected 'Test Site 2', got %s", websites[1][2])
	}
}

func TestInterruptDataConcurrency(t *testing.T) {
	id := &InterruptData{}
	var wg sync.WaitGroup

	// Simulate concurrent additions
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			site := []string{"200", "192.168.1.1", "Site"}
			id.AddWebsite(site)
		}(i)
	}

	wg.Wait()

	websites := id.GetWebsites()
	if len(websites) != 100 {
		t.Errorf("Expected 100 websites, got %d", len(websites))
	}
}

func TestInterruptDataNil(t *testing.T) {
	var id *InterruptData

	// Should not panic
	id.AddWebsite([]string{"200", "192.168.1.1", "Test"})

	websites := id.GetWebsites()
	if websites != nil {
		t.Error("Expected nil websites from nil InterruptData")
	}
}

func TestInterruptDataInitialization(t *testing.T) {
	id := &InterruptData{
		IPBlocks: []string{"192.168.1.0/24"},
		Domain:   "example.com",
		Timeout:  300,
	}

	if len(id.IPBlocks) != 1 {
		t.Error("IPBlocks not initialized correctly")
	}

	if id.Domain != "example.com" {
		t.Error("Domain not initialized correctly")
	}

	if id.Timeout != 300 {
		t.Error("Timeout not initialized correctly")
	}

	if len(id.Websites) != 0 {
		t.Error("Websites should be empty initially")
	}
}

func TestInterruptDataThreadSafety(t *testing.T) {
	id := &InterruptData{}
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id.AddWebsite([]string{"200", "192.168.1.1", "Site"})
		}()
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = id.GetWebsites()
		}()
	}

	wg.Wait()

	// Should not panic and should have 50 websites
	websites := id.GetWebsites()
	if len(websites) != 50 {
		t.Errorf("Expected 50 websites, got %d", len(websites))
	}
}

func TestNewInterruptData(t *testing.T) {
	id := NewInterruptData()

	if id == nil {
		t.Fatal("NewInterruptData returned nil")
	}

	// CancelCh should be initialized (non-nil)
	if id.CancelCh == nil {
		t.Error("CancelCh should be initialized")
	}

	// Should not be cancelled initially
	if id.IsCancelled() {
		t.Error("Should not be cancelled initially")
	}

	// Websites should be nil/empty
	websites := id.GetWebsites()
	if len(websites) != 0 {
		t.Errorf("Websites should be empty, got %d", len(websites))
	}

	// Fields should be zero-valued
	if id.Domain != "" {
		t.Errorf("Domain should be empty, got %s", id.Domain)
	}
	if id.Timeout != 0 {
		t.Errorf("Timeout should be 0, got %d", id.Timeout)
	}
	if len(id.IPBlocks) != 0 {
		t.Errorf("IPBlocks should be empty, got %d", len(id.IPBlocks))
	}
}

func TestCancelAndIsCancelled(t *testing.T) {
	id := NewInterruptData()

	// Before cancel
	if id.IsCancelled() {
		t.Error("Should not be cancelled before Cancel()")
	}

	// Cancel
	id.Cancel()

	// After cancel
	if !id.IsCancelled() {
		t.Error("Should be cancelled after Cancel()")
	}

	// CancelCh should be closed (non-blocking receive)
	select {
	case <-id.CancelCh:
		// OK — channel is closed
	default:
		t.Error("CancelCh should be closed after Cancel()")
	}
}

func TestDoubleCancelSafety(t *testing.T) {
	id := NewInterruptData()

	// First cancel should work
	id.Cancel()
	if !id.IsCancelled() {
		t.Error("Should be cancelled after first Cancel()")
	}

	// Second cancel should NOT panic (double close protection)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Double Cancel() should not panic, got: %v", r)
		}
	}()
	id.Cancel()
}
