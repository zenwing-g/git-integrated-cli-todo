package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"todo/utils"
)

func HandleInit(args []string) {
	const fileName = ".todo.json"

	if _, err := os.Stat(fileName); err == nil {
		fmt.Println("[!] .todo.json already exists.")
		return
	} else if !os.IsNotExist(err) {
		fmt.Println("[-] Error checking file:", err)
		return
	}

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("[-] Failed to create file:", err)
		return
	}
	defer file.Close()

	if len(args) > 0 && args[0] == "--example" {
		tasks := utils.CreateExampleTasks()
		jsonData, _ := json.MarshalIndent(tasks, "", "  ")
		file.Write(jsonData)
		fmt.Println("[+] Example todo list created.")
		return
	}

	file.WriteString("[]")
	fmt.Println("[+] Empty todo list initialized.")
}
