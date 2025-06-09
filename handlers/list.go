package handlers

import (
	"fmt"
	"os"
	"todo/model"
	"todo/storage"

	"github.com/olekukonko/tablewriter"
)

func HandleList(args []string) {
	tasks, err := storage.ReadTasks()
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

	var filtered []model.Task

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
		tmp := []model.Task{}
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
