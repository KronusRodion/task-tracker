package tasks

import (
	"time"

	"github.com/KronusRodion/task-tracker/internal/api/tasks/handler"
	"github.com/KronusRodion/task-tracker/internal/api/tasks/usecase"
	"github.com/KronusRodion/task-tracker/internal/persistence"
	"github.com/KronusRodion/task-tracker/internal/persistence/cache"
	taskhistory "github.com/KronusRodion/task-tracker/internal/persistence/task_history"
	"github.com/KronusRodion/task-tracker/internal/persistence/tasks"
	teammember "github.com/KronusRodion/task-tracker/internal/persistence/team_member"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
)

func NewModule(
	db persistence.TxExecutor,
	redisClient *redis.Client,
	notificationSvc usecase.NotificationService,
	router *mux.Router,
) {
	taskRepo := tasks.NewRepository()
	taskHistoryRepo := taskhistory.NewRepository()
	memberRepo := teammember.NewRepository()

	uow := persistence.NewUnitOfWork(db)
	cache := cache.NewRedisCache(redisClient)

	// Create a new circuit breaker for the notification service
	circuitBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "NotificationService",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	})

	uc := usecase.NewTaskUsecase(
		taskRepo,
		memberRepo,
		taskHistoryRepo,
		uow,
		cache,
		notificationSvc,
		circuitBreaker,
	)
	handler := handler.NewTaskHandler(uc)
	handler.RegisterRoutes(router)
}