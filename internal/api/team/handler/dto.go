package handler

import (
	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/google/uuid"
)

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type TeamResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ListTeamsResponse struct {
	Teams []TeamResponse `json:"teams"`
}

type InviteUserRequest struct {
	UserID uuid.UUID       `json:"user_id"`
	Role   domain.TeamRole `json:"role"` // owner/admin/member
}
