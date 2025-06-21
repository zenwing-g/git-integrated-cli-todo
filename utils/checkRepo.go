package utils

import "os"

// check if this is a git repo
func CheckRepo() bool {
	info, err := os.Stat(".git")
	return err == nil && info.IsDir()
}
