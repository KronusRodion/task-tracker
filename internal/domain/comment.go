package domain

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type TaskComment struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uint64    `json:"task_id"`
	UserID    uuid.UUID `json:"user_id"` 
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *TaskComment) Validate() error {
	if c.TaskID == 0 {
		return errors.New("task_id is required")
	}
	if c.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if c.Content == "" {
		return errors.New("comment is required")
	}
	if utf8.RuneCountInString(c.Content) > 1000 {
		return errors.New("comment is too much")
	}
	return nil
}
