// helpers.go provides shared utility functions for the modules package
package modules

import (
	"html"
	"regexp"
	"strings"
)

// Compiled regex for title extraction (performance optimization)
var titleRegex = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

// ExtractTitle extracts and decodes the title from HTML content
// Returns the decoded title or empty string if not found
func ExtractTitle(htmlContent string) string {
	match := titleRegex.FindStringSubmatch(htmlContent)
	if len(match) > 1 {
		// Decode HTML entities and trim whitespace
		title := html.UnescapeString(match[1])
		title = strings.TrimSpace(title)
		// Remove newlines and excessive whitespace
		title = strings.Join(strings.Fields(title), " ")
		return title
	}
	return ""
}
