package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

type Task struct {
	Name          string     `json:"name"`
	TID           int        `json:"taskid"`
	Description   string     `json:"description"`
	CreatedOnAt   time.Time  `json:"created_on_at"`
	CompletedOnAt *time.Time `json:"completed_on_at"`
	Important     bool       `json:"important"`
	CommandToRun  string     `json:"command_to_run"`
}

var TaskID int = 1

func main() {
	fmt.Println("[*] Starting todo CLI...")

	if len(os.Args) < 2 {
		fmt.Println("[*] Usage:\ntodo [init|ls|add|done|rm]")
		return
	}

	switch os.Args[1] {
	case "init":
		fmt.Println("[*] Running 'init' command")
		handleInit(os.Args[2:])
	case "rm":
		fmt.Println("[*] Running 'rm' command")
		handleDelete()
	case "add":
		fmt.Println("[*] Running 'add' command")
		if err := initTaskID(); err != nil {
			fmt.Println("[x] Failed to initialize taskID:", err)
			return
		}
		handleAdd()
	case "ls":
		fmt.Println("[*] Running 'ls' command")
		handleList(os.Args[2:])
	case "done":
		fmt.Println("[*] Running 'done' command")
		handleCompleted(os.Args[2:])
	default:
		fmt.Println("[x] Unknown command:", os.Args[1])
		fmt.Println("Try: todo [init|ls|add|done|rm]")
	}
}

func handleInit(args []string) {
	const fileName = ".todo.json"
	fmt.Println("[*] Starting 'todo init'...")

	if _, err := os.Stat(fileName); err == nil {
		fmt.Println("[!] .todo.json already exists.")
		return
	} else if !os.IsNotExist(err) {
		fmt.Println("[x] Error checking file:", err)
		return
	}

	fmt.Println("[*] Creating .todo.json file...")
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("[x] Failed to create file:", err)
		return
	}
	defer file.Close()

	if len(args) > 0 && args[0] == "--example" {
		fmt.Println("[*] Writing example tasks...")
		tasks := createExampleTasks()
		jsonData, _ := json.MarshalIndent(tasks, "", "  ")
		file.Write(jsonData)
		fmt.Println("[+] Example todo list created.")
		return
	}

	isRepo := checkRepo()
	fmt.Println("[*] Git repo check:", isRepo)

	if !isRepo {
		fmt.Println("[!] Not a Git repo. Skipping .gitignore.")
	} else {
		fmt.Println("[*] Git repo found. Checking .gitignore...")
		if checkGitignore() {
			fmt.Println("[*] .gitignore exists. Appending entry if missing...")
			content, err := os.ReadFile(".gitignore")
			if err != nil {
				fmt.Println("[x] Failed to read .gitignore:", err)
				return
			}

			if !strings.Contains(string(content), ".todo.json") {
				fmt.Println("[*] Appending .todo.json to .gitignore...")
				f, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Println("[-] Error opening .gitignore:", err)
					return
				}
				defer f.Close()

				_, err = f.WriteString(".todo.json\n")
				if err != nil {
					fmt.Println("[x] Error writing to .gitignore:", err)
					return
				}
			}
		} else {
			fmt.Println("[*] .gitignore not found. Creating it...")
			err := os.WriteFile(".gitignore", []byte("# Created by todo.cli\n.todo.json\n"), 0644)
			if err != nil {
				fmt.Println("[x] Error creating .gitignore:", err)
				return
			}
			fmt.Println("[+] .gitignore created.")
		}
	}

	_, err = file.WriteString("[]")
	if err != nil {
		fmt.Println("[x] Failed to write empty JSON array:", err)
		return
	}
	fmt.Println("[+] Empty todo list initialized.")
}

func handleDelete() {
	const fileName = ".todo.json"
	TaskID = 1
	fmt.Println("[*] Deleting .todo.json...")

	err := os.Remove(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("[!] File does not exist.")
		} else {
			fmt.Println("[x] Could not delete file:", err)
		}
		return
	}

	fmt.Println("[+] .todo.json deleted.")
}

func handleAdd() {
	fmt.Println("[*] Adding new task...")
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

	newTask := Task{
		Name:          name,
		TID:           TaskID,
		Description:   strings.TrimSpace(desc),
		CreatedOnAt:   time.Now(),
		CompletedOnAt: nil,
		Important:     important,
		CommandToRun:  commandToRun,
	}

	tasks, err := readTasks()
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

	if err := os.WriteFile(".todo.json", data, 0644); err != nil {
		fmt.Println("[x] Failed to write file:", err)
		return
	}

	fmt.Println("[+] Task added:", newTask.Name)
}

func handleCompleted(args []string) {
	fmt.Println("[*] Completing task...")
	if len(args) < 1 {
		fmt.Println("[x] No task ID provided")
		return
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("[x] Invalid task ID: ", err)
		return
	}

	tasks, err := readTasks()
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

	if err := os.WriteFile(".todo.json", data, 0644); err != nil {
		fmt.Println("[x] Failed to write tasks:", err)
		return
	}

	fmt.Println("[+] Task marked as completed.")
}

func handleList(args []string) {
	fmt.Println("[*] Listing tasks...")

	tasks, err := readTasks()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File not found. Run 'todo init'.")
			return
		}
		fmt.Println("[x] Failed to read tasks:", err)
		return
	}

	var (
		showCompleted, showUncompleted, filterImportant bool
	)

	for _, arg := range args {
		switch arg {
		case "--com":
			showCompleted = true
		case "--uncom":
			showUncompleted = true
		case "--imp":
			filterImportant = true
		default:
			fmt.Println("[-] Unknown flag:", arg)
		}
	}

	var filtered []Task

	if showCompleted || showUncompleted {
		for _, t := range tasks {
			if showCompleted && t.CompletedOnAt != nil {
				filtered = append(filtered, t)
			}
			if showUncompleted && t.CompletedOnAt == nil {
				filtered = append(filtered, t)
			}
		}
	} else {
		filtered = tasks
	}

	if filterImportant {
		tmp := []Task{}
		for _, t := range filtered {
			if t.Important {
				tmp = append(tmp, t)
			}
		}
		filtered = tmp
	}

	if len(filtered) == 0 {
		fmt.Println("[!] No tasks to show.")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"#", "Name", "Created", "Important"})

	for _, task := range filtered {
		table.Append([]string{
			fmt.Sprintf("%d", task.TID),
			task.Name,
			task.CreatedOnAt.Format("2006-01-02 15:04"),
			fmt.Sprintf("%v", task.Important),
		})
	}
	table.Render()
}

func checkRepo() bool {
	info, err := os.Stat(".git")
	if err != nil {
		return false
	}
	return info.IsDir()
}

func checkGitignore() bool {
	_, err := os.Stat(".gitignore")
	if err != nil {
		return false
	}
	return true
}

func readTasks() ([]Task, error) {
	fmt.Println("[*] Reading tasks from .todo.json")
	data, err := os.ReadFile(".todo.json")
	if err != nil {
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func initTaskID() error {
	fmt.Println("[*] Initializing Task ID")
	tasks, err := readTasks()
	if err != nil {
		if os.IsNotExist(err) {
			TaskID = 1
			return nil
		}
		return err
	}

	maxID := 0
	for _, t := range tasks {
		if t.TID > maxID {
			maxID = t.TID
		}
	}
	TaskID = maxID + 1
	fmt.Println("[*] Task ID set to", TaskID)
	return nil
}

func createExampleTasks() []Task {
	fmt.Println("[*] Creating example tasks...")
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	importants := []bool{false, true}
	completions := []*time.Time{nil, &past}

	var tasks []Task
	count := 1

	for _, imp := range importants {
		for _, comp := range completions {
			tasks = append(tasks, Task{
				Name:          fmt.Sprintf("Task%d", count),
				TID:           TaskID + count - 1,
				Description:   fmt.Sprintf("Description for Task%d", count),
				CreatedOnAt:   now,
				CompletedOnAt: comp,
				Important:     imp,
			})
			count++
		}
	}

	tasks = append(tasks, Task{
		Name:        "BonusTask",
		Description: "This one has no completion time",
		CreatedOnAt: now,
		Important:   false,
		TID:         TaskID + count,
	})

	return tasks
}
