package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/KronusRodion/task-tracker/internal/persistence"

	"github.com/google/uuid"
)

type Repository struct {
}

func NewRepository() Repository {
	return Repository{}
}

func (r Repository) Create(
	ctx context.Context,
	user domain.User,
) error {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return err
	}

	const query = `
	INSERT INTO users (
		id,
		email,
		password,
		full_name,
		created_at,
		updated_at
	)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = exec.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Password,
		user.FullName,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

func (r Repository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.User, error) {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return domain.User{}, err
	}

	const query = `
	SELECT
		id,
		email,
		password,
		full_name,
		created_at,
		updated_at
	FROM users
	WHERE id = ?
	LIMIT 1
	`

	var user domain.User

	err = exec.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FullName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}

		return domain.User{}, err
	}

	return user, nil
}

func (r Repository) GetByEmail(
	ctx context.Context,
	email string,
) (domain.User, error) {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return domain.User{}, err
	}
	const query = `
SELECT
	id,
	email,
	password,
	full_name,
	created_at,
	updated_at
FROM users
WHERE email = ?
LIMIT 1
`

	var user domain.User

	err = exec.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FullName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}

		return domain.User{}, err
	}

	return user, nil
}
