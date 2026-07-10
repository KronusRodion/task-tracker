package teammember

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

func (r Repository) AddMember(
	ctx context.Context,
	teamID, userID uuid.UUID,
	role domain.TeamRole,
) error {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return err
	}
	const query = `
INSERT INTO team_members (
	team_id,
	user_id,
	role,
	joined_at
) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?), ?, NOW())
`

	_, err = exec.ExecContext(ctx, query,
		teamID,
		userID,
		role,
	)

	return err
}

func (r Repository) GetUserRole(
	ctx context.Context,
	teamID, userID uuid.UUID,
) (domain.TeamRole, error) {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return domain.TeamRole(""), err
	}
	const query = `
SELECT role
FROM team_members
WHERE team_id = UUID_TO_BIN(?) AND user_id = UUID_TO_BIN(?)
LIMIT 1
`

	var role domain.TeamRole

	err = exec.QueryRowContext(ctx, query, teamID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrNotTeamMember
		}

		return "", err
	}

	return role, nil
}

func (r Repository) IsMember(
	ctx context.Context,
	teamID, userID uuid.UUID,
) (bool, error) {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return false, err
	}

	const query = `
SELECT 1
FROM team_members
WHERE team_id = UUID_TO_BIN(?) AND user_id = UUID_TO_BIN(?)
LIMIT 1
`

	var tmp int

	err = exec.QueryRowContext(ctx, query, teamID, userID).Scan(&tmp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}