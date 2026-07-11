package handler

import (
	"fmt"
	"time"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/google/uuid"
)

type TaskResponse struct {
	ID          uint64 `json:"id"`
	TeamID      string `json:"team_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`

	CreatedBy  string  `json:"created_by"`
	AssigneeID *string `json:"assignee_id,omitempty"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (t *TaskResponse) FromDomain(task domain.Task) {
	t.ID = task.ID
	t.TeamID = task.TeamID.String()
	t.Title = task.Title
	t.Description = task.Description
	t.Status = string(task.Status)

	t.CreatedBy = task.CreatedBy.String()

	if task.AssigneeID != nil {
		id := task.AssigneeID.String()
		t.AssigneeID = &id
	}

	t.CreatedAt = task.CreatedAt.Format(time.RFC3339)
	t.UpdatedAt = task.UpdatedAt.Format(time.RFC3339)
}

type CreateTaskRequest struct {
	TeamID      string              `json:"team_id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Priority    domain.TaskPriority `json:"priority"`
	AssigneeID  *string             `json:"assignee_id,omitempty"`
}

func (r *CreateTaskRequest) ToDomain(createdBy uuid.UUID) (domain.Task, error) {
	teamID, err := uuid.Parse(r.TeamID)
	if err != nil {
		return domain.Task{}, err
	}

	var assignee *uuid.UUID
	if r.AssigneeID != nil {
		id, err := uuid.Parse(*r.AssigneeID)
		if err != nil {
			return domain.Task{}, err
		}
		assignee = &id
	}

	ok := r.Priority.Validate()
	if !ok {
		return domain.Task{}, fmt.Errorf("invalid priority: '%s'", r.Priority)
	}

	return domain.Task{
		TeamID:      teamID,
		Title:       r.Title,
		Description: r.Description,
		Status:      domain.StatusTodo,
		CreatedBy:   createdBy,
		AssigneeID:  assignee,
		Priority:    r.Priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	AssigneeID  *string `json:"assignee_id,omitempty"`
}

func (r *UpdateTaskRequest) ToDomain() (domain.TaskPatch, error) {

	var patch domain.TaskPatch

	if r.Title != nil {
		patch.Title = r.Title
	}

	if r.Description != nil {
		patch.Description = r.Description
	}

	if r.Status != nil {
		st := domain.TaskStatus(*r.Status)

		if !st.IsValid() {
			return domain.TaskPatch{}, domain.ErrInvalidStatus
		}

		patch.Status = &st
	}

	if r.AssigneeID != nil {
		id, err := uuid.Parse(*r.AssigneeID)
		if err != nil {
			return domain.TaskPatch{}, err
		}
		patch.AssigneeID = &id
	}

	return patch, nil
}

type TaskHistoryResponse struct {
	ID     uint64 `json:"id"`
	TaskID uint64 `json:"task_id"`

	Action string `json:"action"`

	Field    *string `json:"field,omitempty"`
	OldValue *string `json:"old_value,omitempty"`
	NewValue *string `json:"new_value,omitempty"`

	ChangedBy string `json:"changed_by"`
	CreatedAt string `json:"created_at"`
}

func (h *TaskHistoryResponse) FromDomain(hist domain.TaskHistory) {
	h.ID = hist.ID
	h.TaskID = hist.TaskID
	h.Action = string(hist.Action)

	h.Field = hist.Field
	h.OldValue = hist.OldValue
	h.NewValue = hist.NewValue

	h.ChangedBy = hist.ChangedBy.String()
	h.CreatedAt = hist.CreatedAt.Format(time.RFC3339)
}
