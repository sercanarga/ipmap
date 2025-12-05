package modules

import (
	"encoding/json"
	"fmt"
	"ipmap/config"
	"os"
	"strconv"
	"strings"
	"time"
)

type ResultData struct {
	Method          string     `json:"method"`
	SearchSite      string     `json:"search_site,omitempty"`
	Timeout         int        `json:"timeout_ms"`
	IPBlocks        []string   `json:"ip_blocks"`
	FoundedWebsites [][]string `json:"founded_websites"`
	Timestamp       string     `json:"timestamp"`
}

func exportFile(result string, isJSON bool, domain string) {
	ext := ".txt"
	if isJSON {
		ext = ".json"
	}

	// Generate filename based on domain or timestamp
	var fileName string
	if domain != "" {
		// Sanitize domain name for filename (remove special characters)
		safeDomain := strings.ReplaceAll(domain, ".", "_")
		safeDomain = strings.ReplaceAll(safeDomain, "/", "_")
		safeDomain = strings.ReplaceAll(safeDomain, ":", "_")
		fileName = "ipmap_" + safeDomain + "_" + strconv.FormatInt(time.Now().Local().Unix(), 10) + ext
	} else {
		fileName = "ipmap_" + strconv.FormatInt(time.Now().Local().Unix(), 10) + "_export" + ext
	}
	f, err := os.Create(fileName)
	if err != nil {
		config.ErrorLog("Export file creation error: %v", err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(result)
	if err != nil {
		config.ErrorLog("Export file write error: %v", err)
		return
	}

	config.InfoLog("Successfully exported: " + fileName)
}

func PrintResult(method string, title string, timeout int, ipblocks []string, founded [][]string, export bool) {
	fmt.Println()

	// Check if JSON format is requested
	isJSON := config.Format == "json"

	if isJSON {
		// Create JSON result
		result := ResultData{
			Method:          method,
			SearchSite:      title,
			Timeout:         timeout,
			IPBlocks:        ipblocks,
			FoundedWebsites: founded,
			Timestamp:       time.Now().Format(time.RFC3339),
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			config.ErrorLog("JSON marshal error: %v", err)
			return
		}

		fmt.Println(string(jsonData))

		if export {
			exportFile(string(jsonData), true, title)
			return
		}

		fmt.Print("\nDo you want to export result to file? (Y/n): ")
		var ex string
		_, err = fmt.Scanln(&ex)
		if err != nil {
			return
		}

		if ex == "y" || ex == "Y" || ex == "" {
			exportFile(string(jsonData), true, title)
		} else {
			fmt.Println("Export canceled")
		}
	} else {
		// Text format (original)
		resultString := "==================== RESULT ===================="
		resultString += "\nMethod:        " + method

		if title != "" {
			resultString += "\nSearch Site:   " + title
		}

		resultString += "\nTimeout:       " + strconv.Itoa(timeout) + "ms"
		resultString += "\nIP Blocks:     " + strings.Join(ipblocks, ",")

		resultString += "\nFounded Websites:\n"
		if len(founded) > 0 {
			for _, site := range founded {
				// Format: Status, IP, Title[, Hostname]
				if len(site) >= 4 {
					resultString += site[0] + ", " + site[1] + ", " + site[2] + " [" + site[3] + "]\n"
				} else {
					resultString += strings.Join(site, ", ") + "\n"
				}
			}
		}
		resultString += "================================================"
		fmt.Println(resultString)

		if export {
			exportFile(resultString, false, title)
			return
		}

		fmt.Print("\nDo you want to export result to file? (Y/n): ")
		var ex string
		_, err := fmt.Scanln(&ex)
		if err != nil {
			return
		}

		if ex == "y" || ex == "Y" || ex == "" {
			exportFile(resultString, false, title)
		} else {
			fmt.Println("Export canceled")
		}
	}
}
