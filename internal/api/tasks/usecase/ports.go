package usecase

import (
	"context"
	"time"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/google/uuid"
)

// NotificationService defines the interface for sending notifications.
type NotificationService interface {
	SendNotification(ctx context.Context, notification domain.Notification) error
}

// CircuitBreaker defines the interface for a circuit breaker.
type CircuitBreaker interface {
	Execute(fn func() (interface{}, error)) (interface{}, error)
}

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type TaskRepository interface {
	Create(ctx context.Context, task domain.Task) (domain.Task, error)
	GetByID(ctx context.Context, id uint64) (domain.Task, error)
	Update(ctx context.Context, task domain.Task) (domain.Task, error)
	GetByFilter(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, error)
	FindInvalidAssigneeTasks(ctx context.Context) ([]domain.Task, error)
}

type TaskHistoryRepository interface {
	Create(ctx context.Context, history domain.TaskHistory) error
	GetByTaskID(ctx context.Context, taskID uint64) ([]domain.TaskHistory, error)
}

type TeamRepository interface {
	IsMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (bool, error)
}