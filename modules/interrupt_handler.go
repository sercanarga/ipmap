package modules

import "sync"

// InterruptData holds scan data for interrupt handling
type InterruptData struct {
	Websites  [][]string
	IPBlocks  []string
	Domain    string
	Timeout   int
	Cancelled bool          // Flag to indicate cancellation
	CancelCh  chan struct{} // Channel to signal cancellation
	mu        sync.Mutex
}

// NewInterruptData creates a new InterruptData with initialized cancel channel
func NewInterruptData() *InterruptData {
	return &InterruptData{
		CancelCh: make(chan struct{}),
	}
}

// Cancel signals all goroutines to stop
func (id *InterruptData) Cancel() {
	if id == nil {
		return
	}
	id.mu.Lock()
	defer id.mu.Unlock()
	if !id.Cancelled {
		id.Cancelled = true
		close(id.CancelCh)
	}
}

// IsCancelled returns whether the scan has been cancelled
func (id *InterruptData) IsCancelled() bool {
	if id == nil {
		return false
	}
	id.mu.Lock()
	defer id.mu.Unlock()
	return id.Cancelled
}

// AddWebsite safely adds a website to the interrupt data
func (id *InterruptData) AddWebsite(site []string) {
	if id == nil {
		return
	}
	id.mu.Lock()
	defer id.mu.Unlock()
	id.Websites = append(id.Websites, site)
}

// GetWebsites safely retrieves all websites
func (id *InterruptData) GetWebsites() [][]string {
	if id == nil {
		return nil
	}
	id.mu.Lock()
	defer id.mu.Unlock()
	return id.Websites
}
