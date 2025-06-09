package utils

import (
	"fmt"
	"time"
	"todo/model"
	"todo/storage"
)

func CreateExampleTasks() []model.Task {
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	importants := []bool{false, true}
	completions := []*time.Time{nil, &past}

	var tasks []model.Task
	count := 1

	for _, imp := range importants {
		for _, comp := range completions {
			tasks = append(tasks, model.Task{
				Name:          fmt.Sprintf("Task%d", count),
				TID:           storage.TaskID + count - 1,
				Description:   fmt.Sprintf("Description for Task%d", count),
				CreatedOnAt:   now,
				CompletedOnAt: comp,
				Important:     imp,
			})
			count++
		}
	}

	tasks = append(tasks, model.Task{
		Name:        "BonusTask",
		Description: "This one has no completion time",
		CreatedOnAt: now,
		Important:   false,
		TID:         storage.TaskID + count,
	})

	return tasks
}
