package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

type Task struct {
	Name          string     `json:"name"`
	TID           int        `json:"taskid"`
	Description   string     `json:"description"`
	CreatedOnAt   time.Time  `json:"created_on_at"`
	CompletedOnAt *time.Time `json:"completed_on_at,omitempty"`
	Important     bool       `json:"important"`
}

var TaskID int = 1

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:\ntodo [init|ls|add|done|rm]")
		return
	}

	switch os.Args[1] {
	case "init":
		handleInit(os.Args[2:])
	case "rm":
		handleDelete()
	case "add":
		if err := initTaskID(); err != nil {
			fmt.Println("[-] Failed to initialize taskID:", err)
			return
		}
		handleAdd()
	case "ls":
		handleList(os.Args[2:])
	default:
		fmt.Println("Unknown command:", os.Args[1])
		fmt.Println("Try: todo [init|ls|add|done|rm]")
	}
}

func handleInit(args []string) {
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
		tasks := createExampleTasks()
		jsonData, _ := json.MarshalIndent(tasks, "", "  ")
		file.Write(jsonData)
		fmt.Println("[+] Example todo list created.")
		return
	}

	file.WriteString("[]")
	fmt.Println("[+] Empty todo list initialized.")
}

func handleDelete() {
	const fileName = ".todo.json"

	TaskID = 1
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

	newTask := Task{
		Name:          name,
		TID:           TaskID,
		Description:   strings.TrimSpace(desc),
		CreatedOnAt:   time.Now(),
		CompletedOnAt: nil,
		Important:     important,
	}

	tasks, err := readTasks()
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

func handleList(args []string) {
	tasks, err := readTasks()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File not found. Run 'todo init'.")
			return
		}
		fmt.Println("[-] Failed to read tasks:", err)
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

	tasks = append(tasks, Task{
		Name:        "BonusTask",
		Description: "This one has no completion time",
		CreatedOnAt: now,
		Important:   false,
		TID:         TaskID + count,
	})

	return tasks
}
