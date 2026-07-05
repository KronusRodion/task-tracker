package domain

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/KronusRodion/task-tracker/internal/ctxkeys"
	"github.com/google/uuid"
)

var (
	emailRegex      = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	passwordPattern = regexp.MustCompile(`^[a-zA-Z\d\W_]{8,}$`)
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // хешированный пароль
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userCtx, ok := ctx.Value(ctxkeys.UserKey).(UserCtx)
	if !ok {
		return uuid.UUID{}, errors.New("userID was not found")
	}

	return userCtx.ID, nil
}


type UserCtx struct {
	ID uuid.UUID
}

func IsEmailValid(Email string) bool {
	return emailRegex.MatchString(Email) && len([]rune(Email)) < 255
}

func IsPasswordValid(password string) bool {
	return passwordPattern.MatchString(password) && len([]rune(password)) < 255
}
