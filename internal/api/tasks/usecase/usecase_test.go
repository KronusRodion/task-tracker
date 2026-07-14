package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ==================== Mocks for TaskUsecase ====================

type MockTaskRepository struct{ mock.Mock }

func (m *MockTaskRepository) Create(ctx context.Context, task domain.Task) (domain.Task, error) {
	args := m.Called(ctx, task)
	return args.Get(0).(domain.Task), args.Error(1)
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id uint64) (domain.Task, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, task domain.Task) (domain.Task, error) {
	args := m.Called(ctx, task)
	return args.Get(0).(domain.Task), args.Error(1)
}

func (m *MockTaskRepository) GetByFilter(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.Task), args.Error(1)
}

func (m *MockTaskRepository) FindInvalidAssigneeTasks(ctx context.Context) ([]domain.Task, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Task), args.Error(1)
}

type MockTeamRepository struct{ mock.Mock }

func (m *MockTeamRepository) IsMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, teamID, userID)
	return args.Bool(0), args.Error(1)
}

type MockTaskHistoryRepository struct{ mock.Mock }

func (m *MockTaskHistoryRepository) Create(ctx context.Context, history domain.TaskHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockTaskHistoryRepository) GetByTaskID(ctx context.Context, taskID uint64) ([]domain.TaskHistory, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).([]domain.TaskHistory), args.Error(1)
}

type MockUnitOfWork struct{ mock.Mock }

func (m *MockUnitOfWork) DoWithTx(ctx context.Context, fn func(context.Context) error) error {
	args := m.Called(ctx, fn)
	if fn != nil {
		return fn(ctx)
	}
	return args.Error(0)
}

func (m *MockUnitOfWork) Do(ctx context.Context, fn func(context.Context) error) error {
	args := m.Called(ctx, fn)
	if fn != nil {
		return fn(ctx)
	}
	return args.Error(0)
}

type MockCache struct{ mock.Mock }

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// MockCircuitBreaker is a mock implementation of the CircuitBreaker for testing.
type MockCircuitBreaker struct{ mock.Mock }

func (m *MockCircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	args := m.Called(fn)
	return args.Get(0), args.Error(1)
}

// MockNotificationService is a mock implementation of the NotificationService for testing.
type MockNotificationService struct{ mock.Mock }

func (m *MockNotificationService) SendNotification(ctx context.Context, notification domain.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

// ==================== TaskUsecase Suite ====================

type TaskUsecaseSuite struct {
	suite.Suite
	uow             *MockUnitOfWork
	taskRepo        *MockTaskRepository
	teamRepo        *MockTeamRepository
	historyRepo     *MockTaskHistoryRepository
	cache           *MockCache
	notificationSvc *MockNotificationService
	circuitBreaker  *MockCircuitBreaker
	uc              taskUsecase
}

func (s *TaskUsecaseSuite) SetupTest() {
	s.uow = new(MockUnitOfWork)
	s.taskRepo = new(MockTaskRepository)
	s.teamRepo = new(MockTeamRepository)
	s.historyRepo = new(MockTaskHistoryRepository)
	s.cache = new(MockCache)
	s.notificationSvc = new(MockNotificationService)
	s.circuitBreaker = new(MockCircuitBreaker)

	s.uc = NewTaskUsecase(
		s.taskRepo,
		s.teamRepo,
		s.historyRepo,
		s.uow,
		s.cache,
		s.notificationSvc,
		s.circuitBreaker,
	)
}

func TestTaskUsecaseSuite(t *testing.T) {
	suite.Run(t, new(TaskUsecaseSuite))
}

// ====================== CreateTask ======================

func (s *TaskUsecaseSuite) TestCreateTask_Success() {
	task := domain.Task{
		ID:         1,
		TeamID:     uuid.New(),
		CreatedBy:  uuid.New(),
		AssigneeID: nil,
		Title:      "Test Task",
	}

	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teamRepo.On("IsMember", mock.Anything, task.TeamID, task.CreatedBy).Return(true, nil)
	s.taskRepo.On("Create", mock.Anything, task).Return(task, nil)
	s.historyRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	// Настраиваем моки для инвалидации кеша
	for _, status := range domain.AllStatuses() {
		key := fmt.Sprintf("tasks:team:%s:filter:%s", task.TeamID, status)
		s.cache.On("Delete", mock.Anything, key).Return(nil)
	}

	created, err := s.uc.CreateTask(context.Background(), task)
	s.NoError(err)
	s.Equal(task.ID, created.ID)
}

func (s *TaskUsecaseSuite) TestCreateTask_NotTeamMember() {
	task := domain.Task{TeamID: uuid.New(), CreatedBy: uuid.New()}

	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teamRepo.On("IsMember", mock.Anything, task.TeamID, task.CreatedBy).Return(false, nil)

	_, err := s.uc.CreateTask(context.Background(), task)
	s.Assert().Error(err, domain.ErrNotTeamMember)
}

func (s *TaskUsecaseSuite) TestCreateTask_AssigneeNotMember() {
	assigneeID := uuid.New()
	task := domain.Task{
		TeamID:     uuid.New(),
		CreatedBy:  uuid.New(),
		AssigneeID: &assigneeID,
	}

	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teamRepo.On("IsMember", mock.Anything, task.TeamID, task.CreatedBy).Return(true, nil)
	s.teamRepo.On("IsMember", mock.Anything, task.TeamID, assigneeID).Return(false, nil)

	_, err := s.uc.CreateTask(context.Background(), task)
	s.Assert().Error(err, domain.ErrNotTeamMember)
}

// ====================== GetTasks ======================

func (s *TaskUsecaseSuite) TestGetTasks_Success() {
	teamID := uuid.New()
	userID := uuid.New()
	filter := domain.TaskFilter{TeamID: &teamID, UserID: userID}
	tasks := []domain.Task{{ID: 1}, {ID: 2}}

	// Настраиваем моки для вызова Get из кеша (по умолчанию используется StatusTaskFilterAll)
	cacheKey := fmt.Sprintf("tasks:team:%s:filter:%s", teamID, domain.StatusTaskFilterAll)
	s.cache.On("Get", mock.Anything, cacheKey).Return([]byte(nil), redis.Nil)

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teamRepo.On("IsMember", mock.Anything, teamID, userID).Return(true, nil)
	s.taskRepo.On("GetByFilter", mock.Anything, filter).Return(tasks, nil)
	s.cache.On("Set", mock.Anything, cacheKey, mock.Anything, 5*time.Minute).Return(nil)

	result, err := s.uc.GetTasks(context.Background(), filter)
	s.NoError(err)
	s.Len(result, 2)
}

func (s *TaskUsecaseSuite) TestGetTasks_NotTeamMember() {
	teamID := uuid.New()
	userID := uuid.New()
	filter := domain.TaskFilter{TeamID: &teamID, UserID: userID}

	// Настраиваем моки для вызова Get из кеша
	cacheKey := fmt.Sprintf("tasks:team:%s:filter:%s", teamID, domain.StatusTaskFilterAll)
	s.cache.On("Get", mock.Anything, cacheKey).Return([]byte(nil), redis.Nil)

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teamRepo.On("IsMember", mock.Anything, teamID, userID).Return(false, nil)

	_, err := s.uc.GetTasks(context.Background(), filter)
	s.Assert().Error(err, domain.ErrForbidden)
}

func (s *TaskUsecaseSuite) TestGetTasks_CacheHit() {
	teamID := uuid.New()
	userID := uuid.New()
	filter := domain.TaskFilter{TeamID: &teamID, UserID: userID, Status: ptrTo(domain.StatusTodo)}
	tasks := []domain.Task{{ID: 1, Title: "Cached Task"}}

	cacheKey := fmt.Sprintf("tasks:team:%s:filter:%s", teamID, domain.StatusTodo)
	cacheData, _ := json.Marshal(tasks)
	s.cache.On("Get", mock.Anything, cacheKey).Return(cacheData, nil)

	result, err := s.uc.GetTasks(context.Background(), filter)
	s.NoError(err)
	s.Len(result, 1)
	s.Equal("Cached Task", result[0].Title)
}

func (s *TaskUsecaseSuite) TestGetTasks_CacheMiss() {
	teamID := uuid.New()
	userID := uuid.New()
	filter := domain.TaskFilter{TeamID: &teamID, UserID: userID, Status: ptrTo(domain.StatusTodo)}
	tasks := []domain.Task{{ID: 1, Title: "Fresh Task"}}

	cacheKey := fmt.Sprintf("tasks:team:%s:filter:%s", teamID, domain.StatusTodo)
	s.cache.On("Get", mock.Anything, cacheKey).Return([]byte(nil), redis.Nil)

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teamRepo.On("IsMember", mock.Anything, teamID, userID).Return(true, nil)
	s.taskRepo.On("GetByFilter", mock.Anything, filter).Return(tasks, nil)
	s.cache.On("Set", mock.Anything, cacheKey, mock.Anything, 5*time.Minute).Return(nil)

	result, err := s.uc.GetTasks(context.Background(), filter)
	s.NoError(err)
	s.Len(result, 1)
	s.Equal("Fresh Task", result[0].Title)
}

// ====================== UpdateTask ======================

func (s *TaskUsecaseSuite) TestUpdateTask_Success() {
	taskID := uint64(10)
	userID := uuid.New()
	teamID := uuid.New()
	oldTask := domain.Task{ID: taskID, TeamID: teamID, Title: "Old Task"}
	updatedTask := domain.Task{ID: taskID, TeamID: teamID, Title: "New Title"}
	patch := domain.TaskPatch{Title: ptrString("New Title")}

	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.taskRepo.On("GetByID", mock.Anything, taskID).Return(oldTask, nil)
	s.teamRepo.On("IsMember", mock.Anything, teamID, userID).Return(true, nil)
	s.taskRepo.On("Update", mock.Anything, updatedTask).Return(updatedTask, nil)
	s.historyRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	// Настраиваем моки для инвалидации кеша
	for _, status := range domain.AllStatuses() {
		key := fmt.Sprintf("tasks:team:%s:filter:%s", teamID, status)
		s.cache.On("Delete", mock.Anything, key).Return(nil)
	}

	updated, err := s.uc.UpdateTask(context.Background(), taskID, userID, patch)
	s.NoError(err)
	s.Equal("New Title", updated.Title)
}

func (s *TaskUsecaseSuite) TestUpdateTask_TaskNotFound() {
	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.taskRepo.On("GetByID", mock.Anything, mock.Anything).Return(domain.Task{}, domain.ErrTaskNotFound)

	_, err := s.uc.UpdateTask(context.Background(), 999, uuid.New(), domain.TaskPatch{})
	s.Assert().Error(err, domain.ErrTaskNotFound)
}

func (s *TaskUsecaseSuite) TestUpdateTask_Forbidden() {
	task := domain.Task{ID: 5, TeamID: uuid.New()}
	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.taskRepo.On("GetByID", mock.Anything, mock.Anything).Return(task, nil)
	s.teamRepo.On("IsMember", mock.Anything, task.TeamID, mock.Anything).Return(false, nil)

	_, err := s.uc.UpdateTask(context.Background(), 5, uuid.New(), domain.TaskPatch{})
	s.Assert().Error(err, domain.ErrForbidden)
}

// ====================== GetTaskHistory ======================

func (s *TaskUsecaseSuite) TestGetTaskHistory_Success() {
	taskID := uint64(42)
	history := []domain.TaskHistory{{TaskID: taskID}, {TaskID: taskID}}

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.historyRepo.On("GetByTaskID", mock.Anything, taskID).Return(history, nil)

	result, err := s.uc.GetTaskHistory(context.Background(), taskID)
	s.NoError(err)
	s.Len(result, 2)
}

// ====================== FindInvalidAssigneeTasks ======================

func (s *TaskUsecaseSuite) TestFindInvalidAssigneeTasks_Success() {
	teamID := uuid.New()
	assigneeID := uuid.New()
	tasks := []domain.Task{
		{
			ID:         1,
			TeamID:     teamID,
			AssigneeID: &assigneeID,
			Title:      "Invalid Assignee Task",
		},
	}

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.taskRepo.On("FindInvalidAssigneeTasks", mock.Anything).Return(tasks, nil)

	result, err := s.uc.FindInvalidAssigneeTasks(context.Background())
	s.NoError(err)
	s.Len(result, 1)
	s.Equal("Invalid Assignee Task", result[0].Title)
}

func (s *TaskUsecaseSuite) TestFindInvalidAssigneeTasks_Error() {
	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.taskRepo.On("FindInvalidAssigneeTasks", mock.Anything).Return([]domain.Task{}, fmt.Errorf("database error"))

	_, err := s.uc.FindInvalidAssigneeTasks(context.Background())
	s.Error(err)
}

// ====================== Caching Tests ======================

func (s *TaskUsecaseSuite) TestCreateTask_CacheInvalidation() {
	teamID := uuid.New()
	userID := uuid.New()
	task := domain.Task{
		TeamID:    teamID,
		CreatedBy: userID,
		Title:     "New Task",
	}

	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teamRepo.On("IsMember", mock.Anything, teamID, userID).Return(true, nil)
	s.taskRepo.On("Create", mock.Anything, task).Return(task, nil)
	s.historyRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	// Настраиваем моки для инвалидации кеша
	for _, status := range domain.AllStatuses() {
		key := fmt.Sprintf("tasks:team:%s:filter:%s", teamID, status)
		s.cache.On("Delete", mock.Anything, key).Return(nil)
	}

	created, err := s.uc.CreateTask(context.Background(), task)
	s.NoError(err)
	s.Equal(task.Title, created.Title)
}

func (s *TaskUsecaseSuite) TestUpdateTask_CacheInvalidation() {
	taskID := uint64(1)
	teamID := uuid.New()
	userID := uuid.New()
	oldTask := domain.Task{ID: taskID, TeamID: teamID, Title: "Old Task"}
	updatedTask := domain.Task{ID: taskID, TeamID: teamID, Title: "Updated Task"}
	patch := domain.TaskPatch{Title: ptrString("Updated Task")}

	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.taskRepo.On("GetByID", mock.Anything, taskID).Return(oldTask, nil)
	s.teamRepo.On("IsMember", mock.Anything, teamID, userID).Return(true, nil)
	s.taskRepo.On("Update", mock.Anything, updatedTask).Return(updatedTask, nil)
	s.historyRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	// Настраиваем моки для инвалидации кеша
	for _, status := range domain.AllStatuses() {
		key := fmt.Sprintf("tasks:team:%s:filter:%s", teamID, status)
		s.cache.On("Delete", mock.Anything, key).Return(nil)
	}

	updated, err := s.uc.UpdateTask(context.Background(), taskID, userID, patch)
	s.NoError(err)
	s.Equal("Updated Task", updated.Title)
}

// Helper functions
func ptrString(s string) *string {
	return &s
}

func ptrTo(status domain.TaskStatus) *domain.TaskStatus {
	return &status
}
