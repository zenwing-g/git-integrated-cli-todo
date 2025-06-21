package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"todo/constants"
	"todo/utils"
)

func HandleInit(args []string) {
	if _, err := os.Stat(constants.TodoDirPath); err == nil {
		fmt.Println("[!] .todo already exists.")
		return
	} else if !os.IsNotExist(err) {
		fmt.Println("[x] Error checking .todo:", err)
		return
	}

	// create directory
	if err := os.Mkdir(constants.TodoDirPath, 0755); err != nil {
		fmt.Println("[x] Failed to create .todo:", err)
		return
	}
	fmt.Println("[+] Created .todo")

	// create ids.txt
	file, err := os.Create(constants.IdsFilePath)
	if err != nil {
		fmt.Println("[x] Failed to create ids.txt:", err)
		return
	}
	defer file.Close()

	// create .todo.json
	file, err = os.Create(constants.TodoJsonPath)
	if err != nil {
		fmt.Println("[x] Failed to create .todo.json:", err)
		return
	}
	defer file.Close()

	// write example tasks or init empty list
	if len(args) > 0 && args[0] == "--example" {
		tasks := utils.CreateExampleTasks()
		jsonData, _ := json.MarshalIndent(tasks, "", "  ")
		file.Write(jsonData)
		fmt.Println("[+] Example todo list created.")
	} else {
		_, err = file.WriteString("[]")
		if err != nil {
			fmt.Println("[x] Failed to write empty JSON array:", err)
			return
		}
		fmt.Println("[+] Empty todo list initialized.")
	}

	// setup .gitignore only if inside git repo
	if !utils.CheckRepo() {
		fmt.Println("[!] Not a Git repo. Skipping .gitignore.")
		return
	}

	if utils.CheckGitignore() {
		content, err := os.ReadFile(".gitignore")
		if err != nil {
			fmt.Println("[x] Failed to read .gitignore:", err)
			return
		}
		// append entry if not present
		if !strings.Contains(string(content), ".todo.json") {
			f, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println("[-] Error opening .gitignore:", err)
				return
			}
			defer f.Close()
			f.WriteString(".todo.json\n")
		}
	} else {
		// create .gitignore from scratch
		err := os.WriteFile(".gitignore", []byte("# Created by todo.cli\n.todo.json\n"), 0644)
		if err != nil {
			fmt.Println("[x] Error creating .gitignore:", err)
			return
		}
		fmt.Println("[+] .gitignore created.")
	}
}
