package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// check if ID is already used
func CheckIfIDExists(filePath, targetID string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("[x] Failed to open ids.txt:", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == targetID {
			return true
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("[x] Scanner error:", err)
	}
	return false
}
