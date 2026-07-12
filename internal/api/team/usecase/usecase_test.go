package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ==================== Mocks for TeamsUsecase ====================

type MockTeamRepository struct{ mock.Mock }

func (m *MockTeamRepository) Create(ctx context.Context, team domain.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockTeamRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Team, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Team), args.Error(1)
}

func (m *MockTeamRepository) GetUserTeams(ctx context.Context, userID uuid.UUID) ([]domain.Team, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Team), args.Error(1)
}

func (m *MockTeamRepository) GetTeamStats(ctx context.Context, teamID uuid.UUID) (domain.TeamStats, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).(domain.TeamStats), args.Error(1)
}

type MockTeamMemberRepository struct{ mock.Mock }

func (m *MockTeamMemberRepository) AddMember(ctx context.Context, teamID, userID uuid.UUID, role domain.TeamRole) error {
	args := m.Called(ctx, teamID, userID, role)
	return args.Error(0)
}

func (m *MockTeamMemberRepository) GetUserRole(ctx context.Context, teamID, userID uuid.UUID) (domain.TeamRole, error) {
	args := m.Called(ctx, teamID, userID)
	return args.Get(0).(domain.TeamRole), args.Error(1)
}

func (m *MockTeamMemberRepository) IsMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, teamID, userID)
	return args.Bool(0), args.Error(1)
}

type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
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

// ==================== TeamsUsecase Suite ====================

type TeamsUsecaseSuite struct {
	suite.Suite
	uow     *MockUnitOfWork
	teams   *MockTeamRepository
	members *MockTeamMemberRepository
	users   *MockUserRepository
	uc      TeamsUsecase
}

func (s *TeamsUsecaseSuite) SetupTest() {
	s.uow = new(MockUnitOfWork)
	s.teams = new(MockTeamRepository)
	s.members = new(MockTeamMemberRepository)
	s.users = new(MockUserRepository)

	s.uc = New(s.teams, s.members, s.users, s.uow)
}

func TestTeamsUsecaseSuite(t *testing.T) {
	suite.Run(t, new(TeamsUsecaseSuite))
}

// ====================== CreateTeam ======================

func (s *TeamsUsecaseSuite) TestCreateTeam_Success() {
	name := "Development Team"
	ownerID := uuid.New()

	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teams.On("Create", mock.Anything, mock.MatchedBy(func(t domain.Team) bool {
		return t.Name == name && t.CreatedBy == ownerID
	})).Return(nil)
	s.members.On("AddMember", mock.Anything, mock.Anything, ownerID, domain.RoleOwner).Return(nil)

	team, err := s.uc.CreateTeam(context.Background(), name, ownerID)
	s.NoError(err)
	s.Equal(name, team.Name)
	s.NotZero(team.ID)
}

func (s *TeamsUsecaseSuite) TestCreateTeam_CreateFails() {
	name := "Fail Team"
	ownerID := uuid.New()

	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teams.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	_, err := s.uc.CreateTeam(context.Background(), name, ownerID)
	s.Error(err)
}

// ====================== InviteUser ======================

func (s *TeamsUsecaseSuite) TestInviteUser_Success() {
	teamID := uuid.New()
	userID := uuid.New()
	invitedBy := uuid.New()

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teams.On("GetByID", mock.Anything, teamID).Return(domain.Team{}, nil)
	s.members.On("GetUserRole", mock.Anything, teamID, invitedBy).Return(domain.RoleOwner, nil)
	s.users.On("GetByID", mock.Anything, userID).Return(domain.User{}, nil)
	s.members.On("IsMember", mock.Anything, teamID, userID).Return(false, nil)
	s.members.On("AddMember", mock.Anything, teamID, userID, domain.RoleMember).Return(nil)

	err := s.uc.InviteUser(context.Background(), teamID, userID, invitedBy, domain.RoleMember)
	s.NoError(err)
}

func (s *TeamsUsecaseSuite) TestInviteUser_Forbidden() {
	teamID := uuid.New()
	invitedBy := uuid.New()

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teams.On("GetByID", mock.Anything, teamID).Return(domain.Team{}, nil)
	s.members.On("GetUserRole", mock.Anything, teamID, invitedBy).Return(domain.RoleMember, nil)

	err := s.uc.InviteUser(context.Background(), teamID, uuid.New(), invitedBy, domain.RoleMember)
	s.Assert().Error(err, domain.ErrForbidden)
}

func (s *TeamsUsecaseSuite) TestInviteUser_UserAlreadyInTeam() {
	teamID := uuid.New()
	userID := uuid.New()
	invitedBy := uuid.New()

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teams.On("GetByID", mock.Anything, teamID).Return(domain.Team{}, nil)
	s.members.On("GetUserRole", mock.Anything, teamID, invitedBy).Return(domain.RoleOwner, nil)
	s.users.On("GetByID", mock.Anything, userID).Return(domain.User{}, nil)
	s.members.On("IsMember", mock.Anything, teamID, userID).Return(true, nil)

	err := s.uc.InviteUser(context.Background(), teamID, userID, invitedBy, domain.RoleMember)
	s.Error(err)
	s.Contains(err.Error(), "user already in team")
}

func (s *TeamsUsecaseSuite) TestInviteUser_TeamNotFound() {
	teamID := uuid.New()
	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teams.On("GetByID", mock.Anything, teamID).Return(domain.Team{}, domain.ErrTeamNotFound)

	err := s.uc.InviteUser(context.Background(), teamID, uuid.New(), uuid.New(), domain.RoleMember)
	s.Assert().Error(err, domain.ErrTeamNotFound)
}

// ====================== GetUserTeams ======================

func (s *TeamsUsecaseSuite) TestGetUserTeams_Success() {
	userID := uuid.New()
	teams := []domain.Team{
		{ID: uuid.New(), Name: "Team 1"},
		{ID: uuid.New(), Name: "Team 2"},
	}

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teams.On("GetUserTeams", mock.Anything, userID).Return(teams, nil)

	result, err := s.uc.GetUserTeams(context.Background(), userID)
	s.NoError(err)
	s.Len(result, 2)
}

func (s *TeamsUsecaseSuite) TestGetUserTeams_Error() {
	userID := uuid.New()
	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.teams.On("GetUserTeams", mock.Anything, userID).Return([]domain.Team{}, errors.New("db error"))

	_, err := s.uc.GetUserTeams(context.Background(), userID)
	s.Error(err)
}

// ====================== GetTeamStats ======================

func (s *TeamsUsecaseSuite) TestGetTeamStats_Success() {
	teamID := uuid.New()
	userID := uuid.New()
	stats := domain.TeamStats{
		TeamName:       "Test Team",
		MemberCount:    3,
		DoneTasksCount: 10,
	}

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.members.On("IsMember", mock.Anything, teamID, userID).Return(true, nil)
	s.teams.On("GetTeamStats", mock.Anything, teamID).Return(stats, nil)

	result, err := s.uc.GetTeamStats(context.Background(), teamID, userID)
	s.NoError(err)
	s.Equal(stats, result)
}

func (s *TeamsUsecaseSuite) TestGetTeamStats_Error() {
	teamID := uuid.New()
	userID := uuid.New()

	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.members.On("IsMember", mock.Anything, teamID, userID).Return(true, nil)
	s.teams.On("GetTeamStats", mock.Anything, teamID).Return(domain.TeamStats{}, domain.ErrTeamNotFound)

	_, err := s.uc.GetTeamStats(context.Background(), teamID, userID)
	s.Error(err)
	s.Equal(domain.ErrTeamNotFound, err)
}