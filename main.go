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

// Task defines the structure of a single todo item
type Task struct {
	Name          string     `json:"name"`
	TID           int        `json:"taskid"`          // unique Task ID
	Description   string     `json:"description"`     // optional task details
	CreatedOnAt   time.Time  `json:"created_on_at"`   // timestamp of creation
	CompletedOnAt *time.Time `json:"completed_on_at"` // nil if not done
	Important     bool       `json:"important"`       // priority flag
	CommandToRun  string     `json:"command_to_run"`  // optional command to run when marked done
}

// Global counter for new task ID generation
var TaskID int = 1

func main() {
	fmt.Println("[*] Starting todo CLI...")

	// Ensure at least one CLI arg exists
	if len(os.Args) < 2 {
		fmt.Println("[*] Usage:\ntodo [init|ls|add|done|rm]")
		return
	}

	// Handle commands based on first argument
	switch os.Args[1] {
	case "init":
		handleInit(os.Args[2:])
	case "rm":
		handleDelete()
	case "add":
		// Ensure TaskID starts from correct number
		if err := initTaskID(); err != nil {
			fmt.Println("[x] Failed to initialize taskID:", err)
			return
		}
		handleAdd()
	case "ls":
		handleList(os.Args[2:])
	case "done":
		handleCompleted(os.Args[2:])
	default:
		fmt.Println("[x] Unknown command:", os.Args[1])
		fmt.Println("Try: todo [init|ls|add|done|rm]")
	}
}

// Initializes the .todo directory, JSON file, and optionally example tasks
func handleInit(args []string) {
	const dir = ".todo"
	const filePath = ".todo/.todo.json"

	// Check if already initialized
	if _, err := os.Stat(dir); err == nil {
		fmt.Println("[!] .todo already exists.")
		return
	} else if !os.IsNotExist(err) {
		fmt.Println("[x] Error checking .todo:", err)
		return
	}

	// Create .todo directory
	if err := os.Mkdir(dir, 0755); err != nil {
		fmt.Println("[x] Failed to create .todo:", err)
		return
	}
	fmt.Println("[+] Created .todo")

	// Create JSON file
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("[x] Failed to create .todo.json:", err)
		return
	}
	defer file.Close()

	// If "--example" is passed, write sample tasks
	if len(args) > 0 && args[0] == "--example" {
		tasks := createExampleTasks()
		jsonData, _ := json.MarshalIndent(tasks, "", "  ")
		file.Write(jsonData)
		fmt.Println("[+] Example todo list created.")
	} else {
		// Else, just write an empty array
		_, err = file.WriteString("[]")
		if err != nil {
			fmt.Println("[x] Failed to write empty JSON array:", err)
			return
		}
		fmt.Println("[+] Empty todo list initialized.")
	}

	// Git-aware setup for ignoring .todo.json in commits
	if !checkRepo() {
		fmt.Println("[!] Not a Git repo. Skipping .gitignore.")
		return
	}

	// If .gitignore exists, append entry only if missing
	if checkGitignore() {
		content, err := os.ReadFile(".gitignore")
		if err != nil {
			fmt.Println("[x] Failed to read .gitignore:", err)
			return
		}
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
		// Create new .gitignore if missing
		err := os.WriteFile(".gitignore", []byte("# Created by todo.cli\n.todo.json\n"), 0644)
		if err != nil {
			fmt.Println("[x] Error creating .gitignore:", err)
			return
		}
		fmt.Println("[+] .gitignore created.")
	}
}

// Deletes the .todo.json file and resets TaskID
func handleDelete() {
	const fileName = ".todo.json"
	TaskID = 1

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

// Adds a new task via user input and saves to file
func handleAdd() {
	reader := bufio.NewReader(os.Stdin)

	// Task name is mandatory
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

	// Optional description
	fmt.Print("Enter task description (optional): ")
	desc, _ := reader.ReadString('\n')

	// Important flag
	fmt.Print("Is this task important (y/N)? ")
	importantInput, _ := reader.ReadString('\n')
	important := strings.ToLower(strings.TrimSpace(importantInput)) == "y"

	// Optional git command to run on completion
	fmt.Println("Git command to run on completion: ")
	commandInput, _ := reader.ReadString('\n')
	commandToRun := strings.TrimSpace(commandInput)

	// Create new task struct
	newTask := Task{
		Name:          name,
		TID:           TaskID,
		Description:   strings.TrimSpace(desc),
		CreatedOnAt:   time.Now(),
		CompletedOnAt: nil,
		Important:     important,
		CommandToRun:  commandToRun,
	}

	// Read current tasks, append new task
	tasks, err := readTasks()
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("[x] Error reading tasks:", err)
		return
	}
	tasks = append(tasks, newTask)

	// Write updated tasks back to JSON
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

// Marks task as completed, optionally executes a shell command
func handleCompleted(args []string) {
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

	// Look for the task with matching ID
	for i, t := range tasks {
		if t.TID == taskID {
			if t.CompletedOnAt != nil {
				fmt.Println("[!] Task already completed.")
				return
			}
			tasks[i].CompletedOnAt = &now
			found = true

			// Execute post-completion command if available
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

	// Save updated task list
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

// Lists tasks with optional filters: --com, --uncom, --imp
func handleList(args []string) {
	tasks, err := readTasks()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File not found. Run 'todo init'.")
			return
		}
		fmt.Println("[x] Failed to read tasks:", err)
		return
	}

	// Parse flags
	var showCompleted, showUncompleted, filterImportant bool
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

	// Apply filters
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

	// Display as table
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

// Checks if current dir is a Git repo
func checkRepo() bool {
	info, err := os.Stat(".git")
	return err == nil && info.IsDir()
}

// Returns true if .gitignore file exists
func checkGitignore() bool {
	_, err := os.Stat(".gitignore")
	return err == nil
}

// Reads task list from JSON file
func readTasks() ([]Task, error) {
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

// Initializes TaskID by scanning max existing ID
func initTaskID() error {
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
	return nil
}

// Generates sample tasks for --example use case
func createExampleTasks() []Task {
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

	// Add one task without a completion date
	tasks = append(tasks, Task{
		Name:        "BonusTask",
		Description: "This one has no completion time",
		CreatedOnAt: now,
		Important:   false,
		TID:         TaskID + count,
	})

	return tasks
}
