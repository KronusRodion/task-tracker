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
) VALUES (UUID_TO_BIN(?), ?, UUID_TO_BIN(?), ?, ?)
`

	_, err = exec.ExecContext(
		ctx,
		query,
		team.ID,
		team.Name,
		team.CreatedBy,
		team.CreatedAt,
		team.UpdatedAt,
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
	BIN_TO_UUID(id) as id,
	name,
	BIN_TO_UUID(created_by) as created_by,
	created_at,
	updated_at
FROM teams
WHERE id = UUID_TO_BIN(?)
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
WHERE tm.user_id = UUID_TO_BIN(?)
`

	rows, err := exec.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := []domain.Team{}

	for rows.Next() {
		var t domain.Team

		err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.CreatedBy,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		teams = append(teams, t)
	}

	return teams, nil
}

func (r Repository) GetTeamStats(ctx context.Context, teamID uuid.UUID) (domain.TeamStats, error) {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return domain.TeamStats{}, err
	}

query := `
SELECT
	t.name AS team_name,
	COUNT(tm.user_id) AS member_count,
	COUNT(CASE WHEN task.status = 'done' AND task.created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) THEN 1 END) AS done_tasks_count
FROM teams t
LEFT JOIN team_members tm ON t.id = tm.team_id
LEFT JOIN tasks task ON t.id = task.team_id
WHERE t.id = UUID_TO_BIN(?)
GROUP BY t.id, t.name
`
	var stat domain.TeamStats
	err = exec.QueryRowContext(ctx, query, teamID).Scan(&stat.TeamName, &stat.MemberCount, &stat.DoneTasksCount)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.TeamStats{}, domain.ErrTeamNotFound
	}

	if err != nil {
		return domain.TeamStats{}, err
	}

	
	return stat, nil
}