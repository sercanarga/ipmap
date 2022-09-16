package modules

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func exportFile(result string) {
	fileName := "ipmap_" + strconv.FormatInt(time.Now().Local().Unix(), 10) + "_export.txt"
	f, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Export file creation error")
		return
	}
	defer f.Close()

	_, err = f.WriteString(result)
	if err != nil {
		fmt.Println("Export file write error")
		return
	}

	fmt.Println("Successfully exported: " + fileName)
	return

}

func PrintResult(method string, title string, timeout int, ipblocks []string, founded [][]string, export bool) {
	fmt.Println("\n")

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
			resultString += strings.Join(site, ", ") + "\n"
		}
	}
	resultString += "================================================"
	fmt.Println(resultString)

	if export == true {
		exportFile(resultString)
		return
	}

	fmt.Print("\nDo you want to export result to file? (Y/n): ")
	var ex string
	_, err := fmt.Scanln(&ex)
	if err != nil {
		return
	}

	if ex == "y" || ex == "Y" || ex == "" {
		exportFile(resultString)
	} else {
		fmt.Println("Export canceled")
	}
}
