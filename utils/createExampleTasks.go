package utils

import (
	"fmt"
	"time"

	"todo/constants"
)

// generate dummy tasks to populate example list
func CreateExampleTasks() []constants.Task {
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	importants := []bool{false, true}
	completions := []*time.Time{nil, &past}

	var tasks []constants.Task
	count := 1

	for _, imp := range importants {
		for _, comp := range completions {
			tasks = append(tasks, constants.Task{
				Name:          fmt.Sprintf("Task%d", count),
				TID:           GenerateTaskID(8),
				Description:   "Auto-generated task",
				CreatedOnAt:   now,
				CompletedOnAt: comp,
				Important:     imp,
			})
			count++
		}
	}

	tasks = append(tasks, constants.Task{
		Name:        fmt.Sprintf("BonusTask%d", count),
		TID:         GenerateTaskID(8),
		Description: "This one has no completion time",
		CreatedOnAt: now,
		Important:   false,
	})

	return tasks
}
