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
