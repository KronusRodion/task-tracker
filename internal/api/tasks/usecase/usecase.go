// internal/usecase/task_usecase.go
package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/KronusRodion/task-tracker/internal/persistence"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type taskUsecase struct {
	taskRepo    TaskRepository
	teamRepo    TeamRepository
	historyRepo TaskHistoryRepository
	uow         persistence.UnitOfWork
	cache       Cache
}

func NewTaskUsecase(
	taskRepo TaskRepository,
	teamRepo TeamRepository,
	historyRepo TaskHistoryRepository,
	uow persistence.UnitOfWork,
	cache Cache,
) taskUsecase {
	return taskUsecase{
		taskRepo:    taskRepo,
		teamRepo:    teamRepo,
		historyRepo: historyRepo,
		uow:         uow,
		cache:       cache,
	}
}

func (u taskUsecase) CreateTask(ctx context.Context, task domain.Task) (domain.Task, error) {

	return task, u.uow.DoWithTx(ctx, func(ctx context.Context) error {
		ok, err := u.teamRepo.IsMember(ctx, task.TeamID, task.CreatedBy)
		if err != nil {
			return err
		} else if !ok {
			return domain.ErrNotTeamMember
		}

		if task.AssigneeID != nil {
			if ok, err := u.teamRepo.IsMember(ctx, task.TeamID, *task.AssigneeID); err != nil {
				return err
			} else if !ok {
				return domain.ErrNotTeamMember
			}
		}

		task, err = u.taskRepo.Create(ctx, task)
		if err != nil {
			return err
		}

		if err := u.invalidateCache(ctx, task.TeamID); err != nil {
			log.Printf("Failed to invalidate cache: %v", err)
		}

		err = u.historyRepo.Create(ctx, domain.TaskHistory{
			TaskID:    task.ID,
			Action:    domain.HistoryCreated,
			ChangedBy: task.CreatedBy,
			CreatedAt: time.Now(),
		})

		return err
	})
}

func (u taskUsecase) GetTasks(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, error) {
	var tasks []domain.Task
	var err error

	// Формируем ключ для кеша
	status := domain.StatusTaskFilterAll
	if filter.Status != nil {
		status = *filter.Status
	}
	cacheKey := fmt.Sprintf("tasks:team:%s:filter:%s", filter.TeamID, status)

	// Проверяем наличие кеша
	cachedTasks, err := u.getCachedTasks(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	if cachedTasks != nil {
		return cachedTasks, nil
	}

	// Если кеша нет, запрашиваем данные из базы
	err = u.uow.Do(ctx, func(ctx context.Context) error {
		if filter.TeamID != nil {
			if ok, err := u.teamRepo.IsMember(ctx, *filter.TeamID, filter.UserID); err != nil {
				return err
			} else if !ok {
				return domain.ErrForbidden
			}
		}

		tasks, err = u.taskRepo.GetByFilter(ctx, filter)
		return err
	})

	if err != nil {
		return nil, err
	}

	// Сохраняем результат в кеш
	if err := u.cacheTasks(ctx, cacheKey, tasks, 5*time.Minute); err != nil {
		log.Printf("Failed to cache tasks: %v", err)
	}

	return tasks, nil
}

func (u taskUsecase) getCachedTasks(ctx context.Context, key string) ([]domain.Task, error) {
	cachedData, err := u.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var tasks []domain.Task
	if err := json.Unmarshal(cachedData, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (u taskUsecase) cacheTasks(ctx context.Context, key string, tasks []domain.Task, ttl time.Duration) error {
	data, err := json.Marshal(tasks)
	if err != nil {
		return err
	}

	return u.cache.Set(ctx, key, data, ttl)
}

func (u taskUsecase) invalidateCache(ctx context.Context, teamID uuid.UUID) error {
	// Удаляем кеш для всех возможных статусов задач
	for _, status := range domain.AllStatuses() {
		key := fmt.Sprintf("tasks:team:%s:filter:%s", teamID, status)
		if err := u.cache.Delete(ctx, key); err != nil {
			log.Printf("Failed to delete cache key %s: %v", key, err)
		}
	}

	return nil
}

func (u taskUsecase) UpdateTask(
	ctx context.Context,
	taskID uint64,
	userID uuid.UUID,
	patch domain.TaskPatch,
) (domain.Task, error) {

	var result domain.Task

	err := u.uow.DoWithTx(ctx, func(ctx context.Context) error {

		task, err := u.taskRepo.GetByID(ctx, taskID)
		if err != nil {
			if errors.Is(err, domain.ErrTaskNotFound) {
				return domain.ErrTaskNotFound
			}
			return err
		}

		if ok, err := u.teamRepo.IsMember(ctx, task.TeamID, userID); err != nil {
			return err
		} else if !ok {
			return domain.ErrForbidden
		}

		updated := task

		if patch.Title != nil {
			updated.Title = *patch.Title
		}

		if patch.Description != nil {
			updated.Description = *patch.Description
		}

		if patch.Status != nil {
			updated.Status = *patch.Status
		}

		if patch.AssigneeID != nil {
			if ok, err := u.teamRepo.IsMember(ctx, task.TeamID, *patch.AssigneeID); err != nil {
				return err
			} else if !ok {
				return domain.ErrNotTeamMember
			}

			updated.AssigneeID = patch.AssigneeID
		}

		result, err = u.taskRepo.Update(ctx, updated)
		if err != nil {
			return err
		}

		// Инвалидация кеша после обновления задачи
		if err := u.invalidateCache(ctx, task.TeamID); err != nil {
			log.Printf("Failed to invalidate cache: %v", err)
		}

		return u.logChanges(ctx, task, result, userID)
	})

	return result, err
}

func (u taskUsecase) GetTaskHistory(
	ctx context.Context,
	taskID uint64,
) ([]domain.TaskHistory, error) {

	var history []domain.TaskHistory

	err := u.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		history, err = u.historyRepo.GetByTaskID(ctx, taskID)
		return err
	})

	return history, err
}

func (u taskUsecase) FindInvalidAssigneeTasks(ctx context.Context) ([]domain.Task, error) {
	var tasks []domain.Task

	err := u.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		tasks, err = u.taskRepo.FindInvalidAssigneeTasks(ctx)
		return err
	})

	return tasks, err
}

// logChanges - обрабатывает измененные поля задачи и на каждое изменение создает запись в истории
// требует вызова внутри методов unit of work
func (u taskUsecase) logChanges(ctx context.Context, oldTask, newTask domain.Task, changedBy uuid.UUID) error {
	now := time.Now()

	compareAndLog := func(field string, oldVal, newVal *string) error {
		if oldVal == nil && newVal == nil {
			return nil
		}
		if oldVal != nil && newVal != nil && *oldVal == *newVal {
			return nil
		}

		return u.historyRepo.Create(ctx, domain.TaskHistory{
			TaskID:    newTask.ID,
			Action:    domain.HistoryUpdated,
			Field:     &field,
			OldValue:  oldVal,
			NewValue:  newVal,
			ChangedBy: changedBy,
			CreatedAt: now,
		})

	}

	// Title
	err := compareAndLog("title", &oldTask.Title, &newTask.Title)
	if err != nil {
		return err
	}

	// Description
	err = compareAndLog("description", &oldTask.Description, &newTask.Description)
	if err != nil {
		return err
	}

	// Status
	oldStatus := string(oldTask.Status)
	newStatus := string(newTask.Status)
	err = compareAndLog("status", &oldStatus, &newStatus)
	if err != nil {
		return err
	}

	// Assignee
	var oldAssignee, newAssignee *string
	if oldTask.AssigneeID != nil {
		s := oldTask.AssigneeID.String()
		oldAssignee = &s
	}
	if newTask.AssigneeID != nil {
		s := newTask.AssigneeID.String()
		newAssignee = &s
	}
	err = compareAndLog("assignee_id", oldAssignee, newAssignee)
	if err != nil {
		return err
	}
	return nil
}