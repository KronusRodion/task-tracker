package tasks

import (
	"github.com/KronusRodion/task-tracker/internal/api/tasks/handler"
	"github.com/KronusRodion/task-tracker/internal/api/tasks/usecase"
	"github.com/KronusRodion/task-tracker/internal/persistence"
	taskhistory "github.com/KronusRodion/task-tracker/internal/persistence/task_history"
	"github.com/KronusRodion/task-tracker/internal/persistence/tasks"
	teammember "github.com/KronusRodion/task-tracker/internal/persistence/team_member"
	"github.com/gorilla/mux"
)

func NewModule(
	db persistence.TxExecutor,
	router *mux.Router,
) {

	taskRepo := tasks.NewRepository()
	taskHistoryRepo := taskhistory.NewRepository()
	memberRepo := teammember.NewRepository()

	uow := persistence.NewUnitOfWork(db)

	uc := usecase.NewTaskUsecase(taskRepo, memberRepo, taskHistoryRepo, uow)
	handler := handler.NewTaskHandler(uc)

	handler.RegisterRoutes(router)
}
