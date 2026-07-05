package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskHistory struct {
	ID uint64

	TaskID uint64

	Action HistoryAction

	Field *string

	OldValue *string
	NewValue *string

	ChangedBy uuid.UUID

	CreatedAt time.Time
}

type HistoryAction string

const (
	HistoryCreated  HistoryAction = "created"
	HistoryUpdated  HistoryAction = "updated"
	HistoryAssigned HistoryAction = "assigned"
	HistoryStatus   HistoryAction = "status_changed"
	HistoryComment  HistoryAction = "comment_added"
)
