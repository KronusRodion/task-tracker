package auth

import (
	authhandler "github.com/KronusRodion/task-tracker/internal/api/auth/handler"
	"github.com/KronusRodion/task-tracker/internal/api/auth/jwtmanager"
	"github.com/KronusRodion/task-tracker/internal/api/auth/password"
	"github.com/KronusRodion/task-tracker/internal/api/auth/storage"
	"github.com/KronusRodion/task-tracker/internal/api/auth/usecase"
	"github.com/KronusRodion/task-tracker/internal/config"
	"github.com/KronusRodion/task-tracker/internal/persistence"
	userrepo "github.com/KronusRodion/task-tracker/internal/persistence/user"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func NewModule(
	db persistence.TxExecutor,
	cache *redis.Client,
	router *mux.Router,
	cfg config.Auth,
) {

	userRepo := userrepo.NewRepository()

	refreshRepo := storage.New(cache)

	hasher := password.New(0) // default cost

	jwtManager := jwtmanager.New(cfg.AccessSecret, cfg.RefreshSecret, cfg.Issuer, cfg.AccessTTL, cfg.RefreshTTL)
	uow := persistence.NewUnitOfWork(db)

	authUsecase := usecase.New(userRepo, refreshRepo, hasher, jwtManager, uow)

	authHandler := authhandler.New(authUsecase)

	authHandler.RegisterRoutes(router)
}
