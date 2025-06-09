package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"todo/model"
	"todo/storage"
)

func HandleAdd() {
	reader := bufio.NewReader(os.Stdin)

	var name string
	for {
		fmt.Print("Enter task name (required): ")
		input, _ := reader.ReadString('\n')
		name = strings.TrimSpace(input)
		if name != "" {
			break
		}
		fmt.Println("Task name can't be empty.")
	}

	fmt.Print("Enter task description (optional): ")
	desc, _ := reader.ReadString('\n')

	fmt.Print("Is this task important (y/N)? ")
	importantInput, _ := reader.ReadString('\n')
	important := strings.ToLower(strings.TrimSpace(importantInput)) == "y"

	newTask := model.Task{
		Name:          name,
		TID:           storage.TaskID,
		Description:   strings.TrimSpace(desc),
		CreatedOnAt:   time.Now(),
		CompletedOnAt: nil,
		Important:     important,
	}

	tasks, err := storage.ReadTasks()
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("[-] Error reading tasks:", err)
		return
	}
	tasks = append(tasks, newTask)

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		fmt.Println("[-] JSON error:", err)
		return
	}

	if err := os.WriteFile(".todo.json", data, 0644); err != nil {
		fmt.Println("[-] Failed to write file:", err)
		return
	}

	fmt.Println("[+] Task added:", newTask.Name)
}
