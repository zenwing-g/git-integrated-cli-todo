package utils

import (
	"fmt"
	"os"
)

// if ids.txt not there, create it
func CheckIfIDsFileExists(filePath string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create ids.txt: %w", err)
		}
		defer file.Close()
		fmt.Println("[+] ids.txt created")
	} else if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	}
	return nil
}
