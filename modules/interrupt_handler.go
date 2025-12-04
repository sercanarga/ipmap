package modules

import "sync"

// InterruptData holds scan data for interrupt handling
type InterruptData struct {
	Websites [][]string
	IPBlocks []string
	Domain   string
	Timeout  int
	mu       sync.Mutex
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
