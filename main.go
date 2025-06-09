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

var taskID int = 1

func initTaskID() error {
	tasks, err := readTasks()
	if err != nil {
		if os.IsNotExist(err) {
			taskID = 1
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

	taskID = maxID + 1
	return nil
}

// handleInit creates the sacred .todo.json file in your current folder.
// This file is where your precious tasks will live.
// If the file is already there, it won’t destroy your data or mess with it — promise.
// If it trips over any error, it’ll let you know instead of silently dying.
//
// Bonus: If you pass "--example" it creates a fake todo list with all possible task combos.
// Perfect for those who wanna fuck around without writing shit themselves.
func handleInit(args []string) {
	const fileName = ".todo.json"

	if _, err := os.Stat(fileName); err == nil {
		fmt.Println("[!] .todo.json already exists. Chill, your tasks are safe.")
		return
	} else if !os.IsNotExist(err) {
		fmt.Println("[-] Error checking file:", err)
		return
	}

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("[x] Couldn't create .todo.json. Did your computer throw a tantrum?")
		return
	}
	defer file.Close()

	// If user wants example tasks, create a full set, else empty array.
	if len(args) > 0 && args[0] == "--example" {
		tasks := createExampleTasks()
		jsonData, err := json.MarshalIndent(tasks, "", "  ")
		if err != nil {
			fmt.Println("[-] Failed to marshal example tasks:", err)
			return
		}
		_, err = file.Write(jsonData)
		if err != nil {
			fmt.Println("[-] Failed to write example tasks:", err)
			return
		}
		fmt.Println("[+] Example todo list created. Go wild.")
		return
	}

	// Default: empty todo list
	_, err = file.WriteString("[]")
	if err != nil {
		fmt.Println("[-] Failed to initialize .todo.json: Your disk must be full or something.")
		return
	}

	fmt.Println("[+] Congrats! Your todo list is born.")
}

// createExampleTasks returns a list of tasks with every combination of attributes filled.
// So you can test every edge case and show off to new devs how versatile your todo list is.
func createExampleTasks() []Task {
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	// Possible boolean values for Important and CompletedOnAt presence
	importants := []bool{false, true}
	completedTimes := []*time.Time{nil, &past}

	var tasks []Task
	count := 1

	for _, imp := range importants {
		for _, comp := range completedTimes {
			name := fmt.Sprintf("Task%d", count)
			desc := fmt.Sprintf("Description for %s", name)

			task := Task{
				Name:          name,
				TID:           taskID,
				Description:   desc,
				CreatedOnAt:   now,
				CompletedOnAt: comp,
				Important:     imp,
			}
			tasks = append(tasks, task)
			count++
		}
	}

	// Add one extra task for variety
	tasks = append(tasks, Task{
		Name:          fmt.Sprintf("Task%d", count),
		Description:   "This one has no completion time",
		CreatedOnAt:   now,
		CompletedOnAt: nil,
		Important:     false,
	})

	return tasks
}

// handleDelete nukes the .todo.json file from your computer. It's the only way to be sure.
// If the file is missing, it won’t freak out, just tells you.
// Any other failure, it’ll whine about it.
func handleDelete() {
	const fileName = ".todo.json"

	taskID = 1

	err := os.Remove(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("[!] .todo.json doesn’t exist. Nothing to delete, you’re safe.")
		} else {
			fmt.Println("[-] Couldn’t delete .todo.json. Permission issues? Try again.")
		}
		return
	}

	fmt.Println("[+] Deleted .todo.json. Your tasks are officially vapor now.")
}

// Task represents a single todo item — basically your digital sticky note.
// - Name is mandatory, because how else would you remember it?
// - Description is optional for those who like to write essays.
// - CreatedOnAt timestamps when you added it (in case you procrastinate).
// - CompletedOnAt timestamps when you stopped procrastinating, nil if you haven’t.
// - Important flags it as “Hey, do this first!” or “No, seriously, do this now.”
type Task struct {
	Name          string     `json:"name"`
	TID           int        `json:"taskid"`
	Description   string     `json:"description"`
	CreatedOnAt   time.Time  `json:"created_on_at"`
	CompletedOnAt *time.Time `json:"completed_on_at,omitempty"`
	Important     bool       `json:"important"`
}

// handleAdd asks you a bunch of questions like an annoying assistant to get your new task.
// It won’t let you escape without giving a name — deal with it.
// Then it stuffs the new task into your todo file, making sure nothing explodes.
func handleAdd() {
	reader := bufio.NewReader(os.Stdin)

	var name string
	for {
		fmt.Print("Enter task name (required): ")
		nameInput, _ := reader.ReadString('\n')
		name = strings.TrimSpace(nameInput)
		if name != "" {
			break
		}
		fmt.Println("Task name can’t be empty. Try harder.")
	}

	fmt.Print("Enter task description (optional): ")
	descInput, _ := reader.ReadString('\n')
	desc := strings.TrimSpace(descInput)

	fmt.Print("Is this task important(y/N(default)): ")
	importantInput, _ := reader.ReadString('\n')
	important := strings.ToLower(strings.TrimSpace(importantInput)) == "y"

	now := time.Now()

	newTask := Task{
		Name:          name,
		TID:           taskID,
		Description:   desc,
		CreatedOnAt:   now,
		CompletedOnAt: nil,
		Important:     important,
	}

	tasks, err := readTasks()
	if err != nil {
		if os.IsNotExist(err) {
			tasks = []Task{}
		} else {
			fmt.Println("[-] Can’t read tasks: ", err)
			return
		}
	}

	tasks = append(tasks, newTask)

	jsonData, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		fmt.Println("[-] JSON marshalling failed. What did you do?")
		return
	}

	err = os.WriteFile(".todo.json", jsonData, 0644)
	if err != nil {
		fmt.Println("[-] Failed to save tasks. Maybe your disk hates you.")
		return
	}

	fmt.Println("[+] Task added:", newTask.Name)
}

// readTasks tries to read your tasks from .todo.json.
// If the file is missing or corrupted, it’ll let you know.
func readTasks() ([]Task, error) {
	const fileName = ".todo.json"

	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err // Let caller decide what to do.
	}

	var tasks []Task
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		return nil, err // Invalid JSON, someone messed up.
	}

	return tasks, nil
}

// handleList prints your tasks in a nice table, with some filters for picky folks.
// Flags:
// --com   Only show completed tasks (because you love ticking boxes)
// --uncom Only show incomplete tasks (because you’re avoiding work)
// --imp   Only show important tasks (because you want to feel productive)
// Unknown flags will make it grumpy but won’t crash.
func handleList(args []string) {
	tasks, err := readTasks()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File not found. Run 'todo init' and then add some tasks, lazybones.")
			return
		}
		fmt.Println("[-] Failed to read tasks:", err)
		return
	}

	showCompleted := false
	showUncompleted := false
	filterImportant := false

	// Parse your picky filters.
	for _, arg := range args {
		switch arg {
		case "--com":
			showCompleted = true
		case "--uncom":
			showUncompleted = true
		case "--imp":
			filterImportant = true
		default:
			fmt.Println("[-] Unknown flag:", arg, "Try --com, --uncom, or --imp.")
		}
	}

	filtered := []Task{}

	// Filter tasks by completion status.
	for _, t := range tasks {
		if showCompleted && t.CompletedOnAt == nil {
			continue // skip unfinished if you want only finished
		}
		if showUncompleted && t.CompletedOnAt != nil {
			continue // skip finished if you want only unfinished
		}
		filtered = append(filtered, t)
	}

	// If no completion filters, show all.
	if !showCompleted && !showUncompleted {
		filtered = tasks
	}

	// If importance filter set, show only the big deals.
	if filterImportant {
		importantOnly := []Task{}
		for _, t := range filtered {
			if t.Important {
				importantOnly = append(importantOnly, t)
			}
		}
		filtered = importantOnly
	}

	// No tasks? No party.
	if len(filtered) == 0 {
		fmt.Println("[!] Nothing to show. Add some tasks or relax.")
		return
	}

	// Show your tasks in a pretty table because plain lists are boring.
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

// Commands: init, add, ls, rm
// If you forget or mess up, it gently reminds you how to use it.
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
		err := initTaskID()
		if err != nil {
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
