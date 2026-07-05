package usecase

import (
	"context"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/google/uuid"
)

type TaskRepository interface {
	Create(ctx context.Context, task domain.Task) (domain.Task, error)
	GetByID(ctx context.Context, id uint64) (domain.Task, error)
	Update(ctx context.Context, task domain.Task) (domain.Task, error)
	GetByFilter(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, error)
}

type TaskHistoryRepository interface {
	Create(ctx context.Context, history domain.TaskHistory) error
	GetByTaskID(ctx context.Context, taskID uint64) ([]domain.TaskHistory, error)
}

type TeamRepository interface {
	IsMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (bool, error)
}
