package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/google/uuid"
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

// ==================== TaskUsecase Suite ====================

type TaskUsecaseSuite struct {
	suite.Suite
	uow         *MockUnitOfWork
	taskRepo    *MockTaskRepository
	teamRepo    *MockTeamRepository
	historyRepo *MockTaskHistoryRepository
	uc          taskUsecase
}

func (s *TaskUsecaseSuite) SetupTest() {
	s.uow = new(MockUnitOfWork)
	s.taskRepo = new(MockTaskRepository)
	s.teamRepo = new(MockTeamRepository)
	s.historyRepo = new(MockTaskHistoryRepository)

	s.uc = NewTaskUsecase(s.taskRepo, s.teamRepo, s.historyRepo, s.uow)
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
	s.taskRepo.On("Create", mock.Anything, mock.Anything).Return(task, nil)
	s.historyRepo.On("Create", mock.Anything, mock.MatchedBy(func(h domain.TaskHistory) bool {
		return h.Action == domain.HistoryCreated && h.TaskID == task.ID
	})).Return(nil)

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
	filter := domain.TaskFilter{TeamID: nil, UserID: uuid.New()}
	tasks := []domain.Task{{ID: 1}, {ID: 2}}

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.taskRepo.On("GetByFilter", mock.Anything, filter).Return(tasks, nil)

	result, err := s.uc.GetTasks(context.Background(), filter)
	s.NoError(err)
	s.Len(result, 2)
}

func (s *TaskUsecaseSuite) TestGetTasks_NotTeamMember() {
	teamID := uuid.New()
	filter := domain.TaskFilter{TeamID: &teamID, UserID: uuid.New()}

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teamRepo.On("IsMember", mock.Anything, teamID, filter.UserID).Return(false, nil)

	_, err := s.uc.GetTasks(context.Background(), filter)
	s.Assert().Error(err, domain.ErrForbidden)
}

// ====================== UpdateTask ======================

func (s *TaskUsecaseSuite) TestUpdateTask_Success() {
	taskID := uint64(10)
	userID := uuid.New()
	oldTask := domain.Task{ID: taskID, TeamID: uuid.New(), Title: "Old", Status: domain.StatusTodo}
	patch := domain.TaskPatch{Title: ptrString("New Title")}

	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.taskRepo.On("GetByID", mock.Anything, taskID).Return(oldTask, nil)
	s.teamRepo.On("IsMember", mock.Anything, oldTask.TeamID, userID).Return(true, nil)
	s.taskRepo.On("Update", mock.Anything, mock.Anything).Return(oldTask, nil) // можно улучшить
	s.historyRepo.On("Create", mock.Anything, mock.Anything).Return(nil)       // logChanges

	_, err := s.uc.UpdateTask(context.Background(), taskID, userID, patch)
	s.NoError(err)
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

// Helper
func ptrString(s string) *string {
	return &s
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
	s.taskRepo.On("FindInvalidAssigneeTasks", mock.Anything).Return(nil, fmt.Errorf("database error"))

	_, err := s.uc.FindInvalidAssigneeTasks(context.Background())
	s.Error(err)
}
