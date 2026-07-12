package domain

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID uuid.UUID

	Name string

	CreatedBy uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}

type TeamStats struct {
	TeamName       string `json:"team_name"`
	MemberCount    int    `json:"member_count"`
	DoneTasksCount int    `json:"done_tasks_count"`
}
