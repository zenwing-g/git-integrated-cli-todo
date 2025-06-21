package utils

import (
	"fmt"
	"os"
)

// add new ID to ids.txt
func AppendIDToFile(filePath, ID string) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("[x] Could not open ids.txt to append:", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(ID + "\n"); err != nil {
		fmt.Println("[x] Failed to write ID to ids.txt:", err)
	}
}
