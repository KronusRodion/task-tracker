package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/KronusRodion/task-tracker/internal/persistence"
)

type TeamsUsecase struct {
	teams   TeamRepository
	members TeamMemberRepository
	users   UserRepository
	uow     persistence.UnitOfWork
}

func New(
	teams TeamRepository,
	members TeamMemberRepository,
	users UserRepository,
	uow persistence.UnitOfWork,
) TeamsUsecase {
	return TeamsUsecase{
		teams:   teams,
		members: members,
		users:   users,
		uow:     uow,
	}
}

func (u TeamsUsecase) CreateTeam(
	ctx context.Context,
	name string,
	ownerID uuid.UUID,
) (domain.Team, error) {
	var team domain.Team
	err := u.uow.DoWithTx(ctx, func(ctx context.Context) error {
		team = domain.Team{
			ID:        uuid.New(),
			Name:      name,
			CreatedBy: ownerID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := u.teams.Create(ctx, team); err != nil {
			return err
		}

		if err := u.members.AddMember(ctx, team.ID, ownerID, domain.RoleOwner); err != nil {
			return err
		}

		return nil
	})

	return team, err
}

func (u TeamsUsecase) InviteUser(
	ctx context.Context,
	teamID uuid.UUID,
	userID uuid.UUID,
	invitedBy uuid.UUID,
	role domain.TeamRole,
) error {
	return u.uow.Do(ctx, func(ctx context.Context) error {
		_, err := u.teams.GetByID(ctx, teamID)
		if err != nil {
			return err
		}

		inviterRole, err := u.members.GetUserRole(ctx, teamID, invitedBy)
		if err != nil {
			return err
		}

		if inviterRole != domain.RoleOwner && inviterRole != domain.RoleAdmin {
			return domain.ErrForbidden
		}

		_, err = u.users.GetByID(ctx, userID)
		if err != nil {
			return err
		}

		isMember, err := u.members.IsMember(ctx, teamID, userID)
		if err != nil {
			return err
		}

		if isMember {
			return errors.New("user already in team")
		}

		return u.members.AddMember(ctx, teamID, userID, role)
	})
}

func (u TeamsUsecase) GetUserTeams(
	ctx context.Context,
	userID uuid.UUID,
) ([]domain.Team, error) {
	var team []domain.Team
	err := u.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		team, err = u.teams.GetUserTeams(ctx, userID)
		return err
	})
	return team, err
}
