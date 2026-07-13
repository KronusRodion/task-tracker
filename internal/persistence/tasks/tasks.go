package tasks

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/KronusRodion/task-tracker/internal/persistence"
)

type Repository struct{}

func NewRepository() Repository {
	return Repository{}
}

func (r Repository) Create(ctx context.Context, task domain.Task) (domain.Task, error) {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return domain.Task{}, err
	}

	const query = `
INSERT INTO tasks (
	team_id,
	title,
	description,
	status,
	priority,
	created_by,
	assignee_id,
	created_at,
	updated_at
)
VALUES (UUID_TO_BIN(?), ?, ?, ?, ?, UUID_TO_BIN(?), UUID_TO_BIN(?), ?, ?)
`

	result, err := exec.ExecContext(
		ctx,
		query,
		task.TeamID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.CreatedBy,
		task.AssigneeID,
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return domain.Task{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Task{}, err
	}

	task.ID = uint64(id)

	return task, nil
}

func (r Repository) GetByID(ctx context.Context, id uint64) (domain.Task, error) {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return domain.Task{}, err
	}

	const query = `
SELECT
	id,
	BIN_TO_UUID(team_id) as team_id,
	title,
	description,
	status,
	priority,
	BIN_TO_UUID(created_by) as created_by,
	BIN_TO_UUID(assignee_id) as assignee_id,
	created_at,
	updated_at
FROM tasks
WHERE id = ?
LIMIT 1
`

	var task domain.Task

	err = exec.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.TeamID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.CreatedBy,
		&task.AssigneeID,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Task{}, domain.ErrTaskNotFound
		}
		return domain.Task{}, err
	}

	return task, nil
}

func (r Repository) Update(ctx context.Context, task domain.Task) (domain.Task, error) {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return domain.Task{}, err
	}

	const query = `
UPDATE tasks
SET
	title=?,
	description=?,
	status=?,
	priority=?,
	assignee_id=UUID_TO_BIN(?),
	updated_at=?
WHERE id=?
`

	_, err = exec.ExecContext(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.AssigneeID,
		task.UpdatedAt,
		task.ID,
	)

	if err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

func (r Repository) GetByFilter(
	ctx context.Context,
	filter domain.TaskFilter,
) ([]domain.Task, error) {

	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return nil, err
	}

	query := `
SELECT
	id,
	BIN_TO_UUID(team_id) as team_id,
	title,
	description,
	status,
	priority,
	BIN_TO_UUID(created_by) as created_by,
	BIN_TO_UUID(assignee_id) as assignee_id,
	created_at,
	updated_at
FROM tasks
WHERE 1=1
`

	args := make([]any, 0)

	if filter.TeamID != nil {
		query += " AND team_id=UUID_TO_BIN(?)"
		args = append(args, *filter.TeamID)
	}

	if filter.Status != nil {
		query += " AND status=?"
		args = append(args, *filter.Status)
	}

	if filter.AssigneeID != nil {
		query += " AND assignee_id=UUID_TO_BIN(?)"
		args = append(args, *filter.AssigneeID)
	}

	query += `
ORDER BY created_at DESC
LIMIT ?
OFFSET ?
`

	args = append(args, filter.Limit)
	args = append(args, filter.Offset)

	query = strings.TrimSpace(query)

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)

	for rows.Next() {
		var task domain.Task

		if err := rows.Scan(
			&task.ID,
			&task.TeamID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.CreatedBy,
			&task.AssigneeID,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r Repository) FindInvalidAssigneeTasks(ctx context.Context) ([]domain.Task, error) {
	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return nil, err
	}

	const query = `
SELECT
	t.id,
	BIN_TO_UUID(t.team_id) as team_id,
	t.title,
	t.description,
	t.status,
	t.priority,
	BIN_TO_UUID(t.created_by) as created_by,
	BIN_TO_UUID(t.assignee_id) as assignee_id,
	t.created_at,
	t.updated_at
FROM tasks t
WHERE NOT EXISTS (
	SELECT 1
	FROM team_members tm
	WHERE tm.team_id = t.team_id AND tm.user_id = t.assignee_id
)
`

	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)

	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(
			&task.ID,
			&task.TeamID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.CreatedBy,
			&task.AssigneeID,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}