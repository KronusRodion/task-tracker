package team

import (
	"github.com/KronusRodion/task-tracker/internal/api/team/handler"
	"github.com/KronusRodion/task-tracker/internal/api/team/usecase"
	"github.com/KronusRodion/task-tracker/internal/persistence"
	"github.com/KronusRodion/task-tracker/internal/persistence/team"
	teammember "github.com/KronusRodion/task-tracker/internal/persistence/team_member"
	"github.com/KronusRodion/task-tracker/internal/persistence/user"
	"github.com/gorilla/mux"
)

func NewModule(
	db persistence.TxExecutor,
	router *mux.Router,
) {

	teamRepo := team.NewRepository()
	memberRepo := teammember.NewRepository()
	userRepo := user.NewRepository()

	uow := persistence.NewUnitOfWork(db)

	teamsUC := usecase.New(
		teamRepo,
		memberRepo,
		userRepo,
		uow,
	)

	h := handler.New(teamsUC)
	h.RegisterRoutes(router)
}
