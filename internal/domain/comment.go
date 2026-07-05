package domain

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type TaskComment struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TaskID    uuid.UUID `json:"task_id" db:"task_id"` // FK → tasks.id
	UserID    uuid.UUID `json:"user_id" db:"user_id"` // FK → users.id
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (c *TaskComment) Validate() error {
	if c.TaskID == uuid.Nil {
		return errors.New("id is required")
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
