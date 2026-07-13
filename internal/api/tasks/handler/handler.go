package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/KronusRodion/task-tracker/internal/middleware"
	"github.com/KronusRodion/task-tracker/internal/tools/handler"
	"github.com/KronusRodion/task-tracker/internal/tools/parse"
	"github.com/google/uuid"
	gorilla "github.com/gorilla/mux"
)

type TasksUsecase interface {
	CreateTask(ctx context.Context, task domain.Task) (domain.Task, error)
	GetTasks(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, error)
	UpdateTask(ctx context.Context, taskID uint64, userID uuid.UUID, patch domain.TaskPatch) (domain.Task, error)
	GetTaskHistory(ctx context.Context, taskID uint64) ([]domain.TaskHistory, error)
	FindInvalidAssigneeTasks(ctx context.Context) ([]domain.Task, error)
}

type TaskHandler struct {
	usecase TasksUsecase
}

func NewTaskHandler(usecase TasksUsecase) TaskHandler {
	return TaskHandler{usecase: usecase}
}

// RegisterRoutes — регистрирует роуты задач
func (h TaskHandler) RegisterRoutes(r *gorilla.Router) {
	api := r.PathPrefix("/tasks").Subrouter()

	api.Use(middleware.Auth)
	api.Use(middleware.Logger)
	api.HandleFunc("", h.CreateTask).Methods(http.MethodPost)
	api.HandleFunc("", h.GetTasks).Methods(http.MethodGet)
	api.HandleFunc("/{id}", h.UpdateTask).Methods(http.MethodPatch)
	api.HandleFunc("/{id}/history", h.GetTaskHistory).Methods(http.MethodGet)
	api.HandleFunc("/invalid-assignees", h.FindInvalidAssigneeTasks).Methods(http.MethodGet)
}

// @Summary Создать задачу
// @Description Создать новую задачу в команде. Пользователь должен быть участником команды.
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT токен"
// @Param body body CreateTaskRequest true "Данные задачи"
// @Success 201 {object} TaskResponse
// @Failure 400 {object} handler.ErrorResponse "invalid_request или validation_error"
// @Failure 401 {object} handler.ErrorResponse "unauthorized"
// @Failure 403 {object} handler.ErrorResponse "forbidden"
// @Router /api/v1/tasks [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := domain.UserIDFromContext(r.Context())
	if err != nil {
		handler.WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	taskDomain, err := req.ToDomain(userID)
	if err != nil {
		handler.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	created, err := h.usecase.CreateTask(r.Context(), taskDomain)
	if err != nil {
		handler.WriteError(w, http.StatusForbidden, "forbidden", err.Error())
		return
	}

	var resp TaskResponse
	resp.FromDomain(created)
	handler.WriteJSON(w, http.StatusCreated, resp)
}

// @Summary Получить список задач
// @Description Получить задачи с фильтрацией и пагинацией
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT токен"
// @Param team_id query string false "ID команды"
// @Param status query string false "Статус задачи (todo, in_progress, review, done)"
// @Param assignee_id query string false "ID исполнителя"
// @Param limit query int false "Количество записей (max 100)" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} TaskResponse
// @Failure 401 {object} handler.ErrorResponse "unauthorized"
// @Failure 500 {object} handler.ErrorResponse "internal_error"
// @Router /api/v1/tasks [get]
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	userID, err := domain.UserIDFromContext(r.Context())
	if err != nil {
		handler.WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	filter := domain.TaskFilter{
		TeamID:     parse.UUIDPtr(r.URL.Query().Get("team_id")),
		Status:     parse.TaskStatus(r.URL.Query().Get("status")),
		AssigneeID: parse.UUIDPtr(r.URL.Query().Get("assignee_id")),
		UserID:     userID,
	}

	if limit, err := strconv.ParseUint(r.URL.Query().Get("limit"), 10, 64); err == nil && limit > 0 {
		filter.Limit = limit
	} else {
		filter.Limit = 20
	}

	if offset, err := strconv.ParseUint(r.URL.Query().Get("offset"), 10, 64); err == nil {
		filter.Offset = offset
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	tasks, err := h.usecase.GetTasks(r.Context(), filter)
	if err != nil {
		handler.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	resp := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		resp[i].FromDomain(t)
	}

	handler.WriteJSON(w, http.StatusOK, resp)
}

// @Summary Найти задачи с некорректным assignee
// @Description Возвращает список задач, где assignee не является членом команды этой задачи
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT токен"
// @Success 200 {array} TaskResponse
// @Failure 401 {object} handler.ErrorResponse "unauthorized"
// @Failure 500 {object} handler.ErrorResponse "internal_error"
// @Router /api/v1/tasks/invalid-assignees [get]
func (h *TaskHandler) FindInvalidAssigneeTasks(w http.ResponseWriter, r *http.Request) {
	if _, err := domain.UserIDFromContext(r.Context()); err != nil {
		handler.WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	tasks, err := h.usecase.FindInvalidAssigneeTasks(r.Context())
	if err != nil {
		handler.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	resp := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		resp[i].FromDomain(t)
	}

	handler.WriteJSON(w, http.StatusOK, resp)
}

// @Summary Обновить задачу
// @Description Обновить задачу (только для участников команды)
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT токен"
// @Param id path int true "ID задачи"
// @Param body body UpdateTaskRequest true "Данные для обновления"
// @Success 200 {object} TaskResponse
// @Failure 400 {object} handler.ErrorResponse "invalid_request или validation_error"
// @Failure 401 {object} handler.ErrorResponse "unauthorized"
// @Failure 403 {object} handler.ErrorResponse "forbidden"
// @Failure 404 {object} handler.ErrorResponse "not_found"
// @Router /api/v1/tasks/{id} [patch]
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := domain.UserIDFromContext(r.Context())
	if err != nil {
		handler.WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	vars := gorilla.Vars(r)
	taskIDStr := vars["id"]

	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		handler.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid task id")
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	patch, err := req.ToDomain()
	if err != nil {
		handler.WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	updated, err := h.usecase.UpdateTask(r.Context(), taskID, userID, patch)
	if err != nil {
		switch {
		case err == domain.ErrTaskNotFound:
			handler.WriteError(w, http.StatusNotFound, "not_found", err.Error())
		case err == domain.ErrForbidden || err == domain.ErrNotTeamMember:
			handler.WriteError(w, http.StatusForbidden, "forbidden", err.Error())
		default:
			handler.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
		}
		return
	}

	var resp TaskResponse
	resp.FromDomain(updated)
	handler.WriteJSON(w, http.StatusOK, resp)
}

// @Summary Получить историю изменений задачи
// @Description Возвращает историю изменений указанной задачи
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT токен"
// @Param id path int true "ID задачи"
// @Success 200 {array} TaskHistoryResponse
// @Failure 401 {object} handler.ErrorResponse "unauthorized"
// @Failure 500 {object} handler.ErrorResponse "internal_error"
// @Router /api/v1/tasks/{id}/history [get]
func (h *TaskHandler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	if _, err := domain.UserIDFromContext(r.Context()); err != nil {
		handler.WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	vars := gorilla.Vars(r)
	taskIDStr := vars["id"]

	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		handler.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid task id")
		return
	}

	history, err := h.usecase.GetTaskHistory(r.Context(), taskID)
	if err != nil {
		handler.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	resp := make([]TaskHistoryResponse, len(history))
	for i, h := range history {
		resp[i].FromDomain(h)
	}

	handler.WriteJSON(w, http.StatusOK, resp)
}
