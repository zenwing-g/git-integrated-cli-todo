package utils

import (
	"encoding/json"
	"os"

	"todo/constants"
)

// read all tasks from .todo.json
func ReadTasks() ([]constants.Task, error) {
	data, err := os.ReadFile(constants.TodoJsonPath)
	if err != nil {
		return nil, err
	}

	var tasks []constants.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}
