package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"todo/constants"
	"todo/utils"
)

func HandleCompleted(args []string) {
	if len(args) < 1 {
		fmt.Println("[x] No task ID provided")
		return
	}

	taskID := args[0]
	tasks, err := utils.ReadTasks()
	if err != nil {
		fmt.Println("[x] Failed to read tasks:", err)
		return
	}

	now := time.Now()
	found := false

	for i, t := range tasks {
		if t.TID == taskID {
			if t.CompletedOnAt != nil {
				fmt.Println("[!] Task already completed.")
				return
			}
			tasks[i].CompletedOnAt = &now
			found = true

			// run post-task command if available
			if cmd := strings.TrimSpace(t.CommandToRun); cmd != "" {
				fmt.Println("[*] Running command:", cmd)
				out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
				if err != nil {
					fmt.Println("[x] Command failed:", err)
				}
				fmt.Println(string(out))
			}
			break
		}
	}

	if !found {
		fmt.Println("[-] Task ID not found.")
		return
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		fmt.Println("[x] Failed to encode tasks:", err)
		return
	}
	if err := os.WriteFile(constants.TodoJsonPath, data, 0644); err != nil {
		fmt.Println("[x] Failed to write tasks:", err)
		return
	}

	fmt.Println("[+] Task marked as completed.")
}
