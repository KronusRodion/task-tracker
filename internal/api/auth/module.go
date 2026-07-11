package auth

import (
	authhandler "github.com/KronusRodion/task-tracker/internal/api/auth/handler"
	"github.com/KronusRodion/task-tracker/internal/api/auth/jwtmanager"
	"github.com/KronusRodion/task-tracker/internal/api/auth/password"
	"github.com/KronusRodion/task-tracker/internal/api/auth/storage"
	"github.com/KronusRodion/task-tracker/internal/api/auth/usecase"
	"github.com/KronusRodion/task-tracker/internal/config"
	"github.com/KronusRodion/task-tracker/internal/middleware"
	"github.com/KronusRodion/task-tracker/internal/persistence"
	userrepo "github.com/KronusRodion/task-tracker/internal/persistence/user"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

// Запускает модуль аутентификации + инициализирует auth middleware
func NewModule(
	db persistence.TxExecutor,
	cache *redis.Client,
	router *mux.Router,
	cfg config.Auth,
) middleware.Authenticator {

	userRepo := userrepo.NewRepository()

	refreshRepo := storage.New(cache)

	hasher := password.New(0) // default cost

	jwtManager := jwtmanager.New(cfg.AccessSecret, cfg.RefreshSecret, cfg.Issuer, cfg.AccessTTL, cfg.RefreshTTL)
	uow := persistence.NewUnitOfWork(db)

	authUsecase := usecase.New(userRepo, refreshRepo, hasher, jwtManager, uow)
	middleware.InitAuth(authUsecase)
	authHandler := authhandler.New(authUsecase)

	authHandler.RegisterRoutes(router)
	return authUsecase
}
