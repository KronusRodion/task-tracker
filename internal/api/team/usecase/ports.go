package usecase

import (
	"context"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/google/uuid"
)

type TeamRepository interface {
	Create(ctx context.Context, team domain.Team) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Team, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]domain.Team, error)
	GetTeamStats(ctx context.Context, teamID uuid.UUID) (domain.TeamStats, error)
}

type TeamMemberRepository interface {
	AddMember(ctx context.Context, teamID, userID uuid.UUID, role domain.TeamRole) error
	GetUserRole(ctx context.Context, teamID, userID uuid.UUID) (domain.TeamRole, error)
	IsMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}