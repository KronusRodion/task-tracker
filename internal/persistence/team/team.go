package team

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

func (r Repository) Create(ctx context.Context, team domain.Team) error {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return err
	}
	const query = `
INSERT INTO teams (
	id,
	name,
	created_by,
	created_at,
	updated_at
) VALUES (?, ?, ?, ?)
`

	_, err = exec.ExecContext(
		ctx,
		query,
		team.ID,
		team.Name,
		team.CreatedBy,
		team.CreatedAt,
	)

	return err
}

func (r Repository) GetByID(ctx context.Context, id uuid.UUID) (domain.Team, error) {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return domain.Team{}, err
	}

	const query = `
SELECT
	id,
	name,
	created_by,
	created_at,
	updated_at
FROM teams
WHERE id = ?
LIMIT 1
`

	var t domain.Team

	err = exec.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.Name,
		&t.CreatedBy,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Team{}, domain.ErrTeamNotFound
		}
		return domain.Team{}, err
	}

	return t, nil
}

func (r Repository) GetUserTeams(ctx context.Context, userID uuid.UUID) ([]domain.Team, error) {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return nil, err
	}

	const query = `
SELECT
	t.id,
	t.name,
	t.created_by,
	t.created_at,
	t.updated_at
FROM teams t
JOIN team_members tm ON tm.team_id = t.id
WHERE tm.user_id = ?
`

	rows, err := exec.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := []domain.Team{}

	for rows.Next() {
		var t domain.Team

		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.CreatedBy,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}

		teams = append(teams, t)
	}

	return teams, nil
}
