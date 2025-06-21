package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"todo/constants"
	"todo/utils"
)

// collect task data from user and store it
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

	fmt.Println("Git command to run on completion: ")
	commandInput, _ := reader.ReadString('\n')
	commandToRun := strings.TrimSpace(commandInput)

	taskID := utils.GenerateTaskID(8)

	newTask := constants.Task{
		Name:          name,
		TID:           taskID,
		Description:   strings.TrimSpace(desc),
		CreatedOnAt:   time.Now(),
		CompletedOnAt: nil,
		Important:     important,
		CommandToRun:  commandToRun,
	}

	tasks, err := utils.ReadTasks()
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("[x] Error reading tasks:", err)
		return
	}
	tasks = append(tasks, newTask)

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		fmt.Println("[x] JSON error:", err)
		return
	}
	if err := os.WriteFile(constants.TodoJsonPath, data, 0644); err != nil {
		fmt.Println("[x] Failed to write file:", err)
		return
	}

	fmt.Println("[+] Task added:", newTask.Name)
}
