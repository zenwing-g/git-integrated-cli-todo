package handlers

import (
	"fmt"
	"os"
	"todo/constants"
	"todo/utils"

	"github.com/olekukonko/tablewriter"
)

func HandleList(args []string) {
	tasks, err := utils.ReadTasks()
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
			return
		}
	}

	// apply filters
	filtered := []constants.Task{}
	for _, t := range tasks {
		if showCompleted && t.CompletedOnAt != nil {
			filtered = append(filtered, t)
		} else if showUncompleted && t.CompletedOnAt == nil {
			filtered = append(filtered, t)
		} else if !showCompleted && !showUncompleted {
			filtered = append(filtered, t)
		}
	}

	if filterImportant {
		tmp := []constants.Task{}
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
