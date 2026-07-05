package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TeamMember struct {
	TeamID    uuid.UUID `json:"team_id" db:"team_id"` // FK → teams.id
	UserID    uuid.UUID `json:"user_id" db:"user_id"` // FK → users.id
	Role      TeamRole  `json:"role" db:"role"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (tm *TeamMember) Validate() error {
	if tm.TeamID == uuid.Nil {
		return errors.New("team_id is required")
	}
	if tm.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if !tm.Role.Validate() {
		return errors.New("ivalid role")
	}
	return nil
}
