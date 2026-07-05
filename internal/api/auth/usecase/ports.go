package usecase

import (
	"context"
	"time"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error

	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)

	GetByEmail(ctx context.Context, email string) (domain.User, error)
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, tokenID string, ttl time.Duration) error
	Delete(ctx context.Context, tokenID string) error
	Consume(ctx context.Context, tokenID string) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type JWTManager interface {
	CreateAccess(user *domain.User) (string, error)
	CreateRefresh(user *domain.User) (string, string, time.Time, error)

	ParseAccess(token string) (*domain.AccessClaims, error)
	ParseRefresh(token string) (*domain.RefreshClaims, error)
}
