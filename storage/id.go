package storage

import (
	"os"
)

var TaskID int = 1

func InitTaskID() error {
	tasks, err := ReadTasks()
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
