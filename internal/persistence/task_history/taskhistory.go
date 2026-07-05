package taskhistory

import (
	"context"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/KronusRodion/task-tracker/internal/persistence"
)

type Repository struct{}

func NewRepository() Repository {
	return Repository{}
}

func (r Repository) Create(
	ctx context.Context,
	h domain.TaskHistory,
) error {

	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return err
	}

	const query = `
INSERT INTO task_history(
	task_id,
	action,
	field,
	old_value,
	new_value,
	changed_by,
	created_at
)
VALUES (?, ?, ?, ?, ?, ?, ?)
`

	_, err = exec.ExecContext(
		ctx,
		query,
		h.TaskID,
		h.Action,
		h.Field,
		h.OldValue,
		h.NewValue,
		h.ChangedBy,
		h.CreatedAt,
	)

	return err
}

func (r Repository) GetByTaskID(
	ctx context.Context,
	taskID uint64,
) ([]domain.TaskHistory, error) {

	exec, err := persistence.GetExec(ctx)
	if err != nil {
		return nil, err
	}

	const query = `
SELECT
	id,
	task_id,
	action,
	field,
	old_value,
	new_value,
	changed_by,
	created_at
FROM task_history
WHERE task_id=?
ORDER BY created_at
`

	rows, err := exec.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]domain.TaskHistory, 0)

	for rows.Next() {
		var h domain.TaskHistory

		if err := rows.Scan(
			&h.ID,
			&h.TaskID,
			&h.Action,
			&h.Field,
			&h.OldValue,
			&h.NewValue,
			&h.ChangedBy,
			&h.CreatedAt,
		); err != nil {
			return nil, err
		}

		history = append(history, h)
	}

	return history, rows.Err()
}
