package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusReview     TaskStatus = "review"
	StatusDone       TaskStatus = "done"
)

func (s TaskStatus) IsValid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusReview, StatusDone:
		return true
	default:
		return false
	}
}

type TaskPriority string

func (t TaskPriority) Validate() bool {
	return t == PriorityLow || t == PriorityMedium || t == PriorityHigh
}

const (
	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"
)

type Task struct {
	ID uint64

	TeamID uuid.UUID

	Title       string
	Description string
	Priority    TaskPriority

	Status TaskStatus

	CreatedBy  uuid.UUID
	AssigneeID *uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}

type TaskFilter struct {
	TeamID     *uuid.UUID
	UserID     uuid.UUID
	Status     *TaskStatus
	AssigneeID *uuid.UUID

	Limit  uint64
	Offset uint64
}

type TaskPatch struct {
	Title       *string     `json:"title,omitempty"`
	Description *string     `json:"description,omitempty"`
	Status      *TaskStatus `json:"status,omitempty"`
	AssigneeID  *uuid.UUID  `json:"assignee_id,omitempty"`
}



// TopUser represents a user ranked by task count
type TopUser struct {
	UserID     uuid.UUID `json:"user_id"`
	Username   string    `json:"username"`
	TaskCount  int       `json:"task_count"`
	Rank       int       `json:"rank"`
}
