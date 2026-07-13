package tasks

import (
	"github.com/KronusRodion/task-tracker/internal/api/tasks/handler"
	"github.com/KronusRodion/task-tracker/internal/api/tasks/usecase"
	"github.com/KronusRodion/task-tracker/internal/persistence"
	"github.com/KronusRodion/task-tracker/internal/persistence/cache"
	taskhistory "github.com/KronusRodion/task-tracker/internal/persistence/task_history"
	"github.com/KronusRodion/task-tracker/internal/persistence/tasks"
	teammember "github.com/KronusRodion/task-tracker/internal/persistence/team_member"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func NewModule(
	db persistence.TxExecutor,
	redisClient *redis.Client,
	router *mux.Router,
) {

	taskRepo := tasks.NewRepository()
	taskHistoryRepo := taskhistory.NewRepository()
	memberRepo := teammember.NewRepository()

	uow := persistence.NewUnitOfWork(db)
	cache := cache.NewRedisCache(redisClient)

	uc := usecase.NewTaskUsecase(taskRepo, memberRepo, taskHistoryRepo, uow, cache)
	handler := handler.NewTaskHandler(uc)

	handler.RegisterRoutes(router)
}
