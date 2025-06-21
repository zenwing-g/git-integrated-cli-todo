package constants

import "time"

// data model for each task in the list
type Task struct {
	Name          string     `json:"name"`
	TID           string     `json:"taskid"`
	Description   string     `json:"description"`
	CreatedOnAt   time.Time  `json:"created_on_at"`
	CompletedOnAt *time.Time `json:"completed_on_at"`
	Important     bool       `json:"important"`
	CommandToRun  string     `json:"command_to_run"`
}
