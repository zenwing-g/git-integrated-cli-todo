package handlers

import (
	"fmt"
	"os"

	"todo/constants"
)

// nukes the .todo dir
func HandleDelete() {
	err := os.RemoveAll(constants.TodoDirPath)
	if err != nil {
		fmt.Println("[x] Failed to delete directory: ", err)
	} else {
		fmt.Println("[-] .todo removed")
	}
}
