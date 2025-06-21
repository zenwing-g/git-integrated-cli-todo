package utils

import "os"

// check if .gitignore exists
func CheckGitignore() bool {
	_, err := os.Stat(".gitignore")
	return err == nil
}
