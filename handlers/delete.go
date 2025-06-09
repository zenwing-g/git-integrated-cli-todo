package handlers

import (
	"fmt"
	"os"
	"todo/storage"
)

func HandleDelete() {
	const fileName = ".todo.json"

	storage.TaskID = 1
	err := os.Remove(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("[!] File does not exist.")
		} else {
			fmt.Println("[-] Could not delete file:", err)
		}
		return
	}

	fmt.Println("[+] .todo.json deleted.")
}
