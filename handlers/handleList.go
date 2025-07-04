package handlers

import (
	"fmt"
	"os"
	"strings"
	"todo/constants"
	"todo/utils"
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

	// Add "Status" column
	headers := []string{"ID", "Name", "Created", "Important", "Status"}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}

	for _, task := range filtered {
		if len(task.TID) > colWidths[0] {
			colWidths[0] = len(task.TID)
		}
		if len(task.Name) > colWidths[1] {
			colWidths[1] = len(task.Name)
		}
		created := task.CreatedOnAt.Format("2006-01-02 15:04")
		if len(created) > colWidths[2] {
			colWidths[2] = len(created)
		}
		imp := fmt.Sprintf("%v", task.Important)
		if len(imp) > colWidths[3] {
			colWidths[3] = len(imp)
		}
		status := "Pending"
		if task.CompletedOnAt != nil {
			status = "Done"
		}
		if len(status) > colWidths[4] {
			colWidths[4] = len(status)
		}
	}

	// Helper: build horizontal border
	buildBorder := func() string {
		parts := []string{"+"}
		for _, w := range colWidths {
			parts = append(parts, strings.Repeat("-", w+2), "+")
		}
		return strings.Join(parts, "")
	}

	// Helper: build row
	buildRow := func(cols []string) string {
		parts := []string{"|"}
		for i, col := range cols {
			padding := colWidths[i] - len(col)
			parts = append(parts, " "+col+strings.Repeat(" ", padding+1), "|")
		}
		return strings.Join(parts, "")
	}

	// Print table
	fmt.Println(buildBorder())
	fmt.Println(buildRow(headers))
	fmt.Println(buildBorder())

	for _, task := range filtered {
		status := "Pending"
		if task.CompletedOnAt != nil {
			status = "Done"
		}
		row := []string{
			task.TID,
			task.Name,
			task.CreatedOnAt.Format("2006-01-02 15:04"),
			fmt.Sprintf("%v", task.Important),
			status,
		}
		fmt.Println(buildRow(row))
	}
	fmt.Println(buildBorder())
}
