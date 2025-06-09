package model

import "time"

type Task struct {
	Name          string     `json:"name"`
	TID           int        `json:"taskid"`
	Description   string     `json:"description"`
	CreatedOnAt   time.Time  `json:"created_on_at"`
	CompletedOnAt *time.Time `json:"completed_on_at,omitempty"`
	Important     bool       `json:"important"`
}
