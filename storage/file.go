package storage

import (
	"encoding/json"
	"os"
	"todo/model"
)

func ReadTasks() ([]model.Task, error) {
	data, err := os.ReadFile(".todo.json")
	if err != nil {
		return nil, err
	}

	var tasks []model.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}
