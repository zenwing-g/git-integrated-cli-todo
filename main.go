package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

// data model for each task in the list
type Task struct {
	Name          string     `json:"name"`
	TID           string     `json:"taskid"`
	Description   string     `json:"description"`
	CreatedOnAt   time.Time  `json:"created_on_at"`
	CompletedOnAt *time.Time `json:"completed_on_at"`
	Important     bool       `json:"important"`
	CommandToRun  string     `json:"command_to_run"`
}

// constants for file paths and task ID generation
const (
	charSet      = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+-="
	TodoDirPath  = ".todo"
	TodoJsonPath = ".todo/.todo.json"
	IdsFilePath  = ".todo/ids.txt"
)

func main() {
	fmt.Println("[*] Starting todo CLI...")

	// basic CLI routing
	if len(os.Args) < 2 {
		fmt.Println("[*] Usage:\ntodo [init|ls|add|done|rm]")
		return
	}

	switch os.Args[1] {
	case "init":
		handleInit(os.Args[2:])
	case "rm":
		handleDelete()
	case "add":
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

// sets up .todo dir and files
func handleInit(args []string) {
	if _, err := os.Stat(TodoDirPath); err == nil {
		fmt.Println("[!] .todo already exists.")
		return
	} else if !os.IsNotExist(err) {
		fmt.Println("[x] Error checking .todo:", err)
		return
	}

	// create directory
	if err := os.Mkdir(TodoDirPath, 0755); err != nil {
		fmt.Println("[x] Failed to create .todo:", err)
		return
	}
	fmt.Println("[+] Created .todo")

	// create ids.txt
	file, err := os.Create(IdsFilePath)
	if err != nil {
		fmt.Println("[x] Failed to create ids.txt:", err)
		return
	}
	defer file.Close()

	// create .todo.json
	file, err = os.Create(TodoJsonPath)
	if err != nil {
		fmt.Println("[x] Failed to create .todo.json:", err)
		return
	}
	defer file.Close()

	// write example tasks or init empty list
	if len(args) > 0 && args[0] == "--example" {
		tasks := createExampleTasks()
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
	if !checkRepo() {
		fmt.Println("[!] Not a Git repo. Skipping .gitignore.")
		return
	}

	if checkGitignore() {
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

// nukes the .todo dir
func handleDelete() {
	err := os.RemoveAll(TodoDirPath)
	if err != nil {
		fmt.Println("[x] Failed to delete directory: ", err)
	} else {
		fmt.Println("[-] .todo removed")
	}
}

// collect task data from user and store it
func handleAdd() {
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

	taskID := generateTaskID(8)

	newTask := Task{
		Name:          name,
		TID:           taskID,
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
	if err := os.WriteFile(TodoJsonPath, data, 0644); err != nil {
		fmt.Println("[x] Failed to write file:", err)
		return
	}

	fmt.Println("[+] Task added:", newTask.Name)
}

// marks task as completed and optionally runs shell command
func handleCompleted(args []string) {
	if len(args) < 1 {
		fmt.Println("[x] No task ID provided")
		return
	}

	taskID := args[0]
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
	if err := os.WriteFile(TodoJsonPath, data, 0644); err != nil {
		fmt.Println("[x] Failed to write tasks:", err)
		return
	}

	fmt.Println("[+] Task marked as completed.")
}

// show tasks with filter options
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

	// handle flags
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

	// apply filters
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

	// print as table
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"ID", "Name", "Created", "Important"})

	for _, task := range filtered {
		table.Append([]string{
			task.TID,
			task.Name,
			task.CreatedOnAt.Format("2006-01-02 15:04"),
			fmt.Sprintf("%v", task.Important),
		})
	}
	table.Render()
}

// read all tasks from .todo.json
func readTasks() ([]Task, error) {
	data, err := os.ReadFile(TodoJsonPath)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// generate unique ID, avoid collisions
func generateTaskID(length int) string {
	filePath := IdsFilePath

	for {
		id := make([]byte, length)
		max := big.NewInt(int64(len(charSet)))

		for i := range id {
			n, err := rand.Int(rand.Reader, max)
			if err != nil {
				panic(err)
			}
			id[i] = charSet[n.Int64()]
		}
		ID := string(id)

		if err := checkIfIdsFileExists(filePath); err != nil {
			fmt.Println("[x] Error ensuring ids.txt exists:", err)
			return ""
		}

		if !checkIfIdExists(filePath, ID) {
			appendIDToFile(filePath, ID)
			return ID
		}
	}
}

// check if ID is already used
func checkIfIdExists(filePath, targetID string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("[x] Failed to open ids.txt:", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == targetID {
			return true
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("[x] Scanner error:", err)
	}
	return false
}

// if ids.txt not there, create it
func checkIfIdsFileExists(filePath string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create ids.txt: %w", err)
		}
		defer file.Close()
		fmt.Println("[+] ids.txt created")
	} else if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	}
	return nil
}

// add new ID to ids.txt
func appendIDToFile(filePath, ID string) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("[x] Could not open ids.txt to append:", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(ID + "\n"); err != nil {
		fmt.Println("[x] Failed to write ID to ids.txt:", err)
	}
}

// check if this is a git repo
func checkRepo() bool {
	info, err := os.Stat(".git")
	return err == nil && info.IsDir()
}

// check if .gitignore exists
func checkGitignore() bool {
	_, err := os.Stat(".gitignore")
	return err == nil
}

// generate dummy tasks to populate example list
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
				TID:           generateTaskID(8),
				Description:   "Auto-generated task",
				CreatedOnAt:   now,
				CompletedOnAt: comp,
				Important:     imp,
			})
			count++
		}
	}

	tasks = append(tasks, Task{
		Name:        fmt.Sprintf("BonusTask%d", count),
		TID:         generateTaskID(8),
		Description: "This one has no completion time",
		CreatedOnAt: now,
		Important:   false,
	})

	return tasks
}
