package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ==================== Mocks ====================

type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type MockRefreshTokenRepository struct{ mock.Mock }

func (m *MockRefreshTokenRepository) Save(ctx context.Context, tokenID string, ttl time.Duration) error {
	args := m.Called(ctx, tokenID, ttl)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) Consume(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) Delete(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

type MockPasswordHasher struct{ mock.Mock }

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) Compare(hashed, password string) error {
	args := m.Called(hashed, password)
	return args.Error(0)
}

type MockJWTManager struct{ mock.Mock }

func (m *MockJWTManager) CreateAccess(user *domain.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) CreateRefresh(user *domain.User) (string, string, time.Time, error) {
	args := m.Called(user)
	return args.String(0), args.String(1), args.Get(2).(time.Time), args.Error(3)
}

func (m *MockJWTManager) ParseRefresh(token string) (*domain.RefreshClaims, error) {
	args := m.Called(token)
	if c := args.Get(0); c != nil {
		return c.(*domain.RefreshClaims), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockJWTManager) ParseAccess(token string) (*domain.AccessClaims, error) {
	args := m.Called(token)
	if c := args.Get(0); c != nil {
		return c.(*domain.AccessClaims), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockUnitOfWork struct{ mock.Mock }

func (m *MockUnitOfWork) DoWithTx(ctx context.Context, fn func(context.Context) error) error {
	args := m.Called(ctx, fn)
	// Выполняем переданную функцию (имитируем транзакцию)
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

// ==================== AuthUsecase Suite ====================

type AuthUsecaseSuite struct {
	suite.Suite
	uow     *MockUnitOfWork
	users   *MockUserRepository
	refresh *MockRefreshTokenRepository
	hasher  *MockPasswordHasher
	jwt     *MockJWTManager
	uc      AuthUsecase
}

func (s *AuthUsecaseSuite) SetupTest() {
	s.uow = new(MockUnitOfWork)
	s.users = new(MockUserRepository)
	s.refresh = new(MockRefreshTokenRepository)
	s.hasher = new(MockPasswordHasher)
	s.jwt = new(MockJWTManager)

	s.uc = New(s.users, s.refresh, s.hasher, s.jwt, s.uow)
}

func TestAuthUsecaseSuite(t *testing.T) {
	suite.Run(t, new(AuthUsecaseSuite))
}

// ====================== Register ======================

func (s *AuthUsecaseSuite) TestRegister_Success() {
	email := "test@example.com"
	password := "pass123"
	fullName := "Test User"

	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.users.On("GetByEmail", mock.Anything, email).Return(domain.User{}, domain.ErrUserNotFound)
	s.hasher.On("Hash", password).Return("hashedpass", nil)
	s.users.On("Create", mock.Anything, mock.MatchedBy(func(u domain.User) bool {
		return u.Email == email && u.FullName == fullName
	})).Return(nil)

	err := s.uc.Register(context.Background(), email, password, fullName)
	s.NoError(err)
}

func (s *AuthUsecaseSuite) TestRegister_UserAlreadyExists() {
	email := "exists@example.com"
	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.users.On("GetByEmail", mock.Anything, email).Return(domain.User{}, nil)

	err := s.uc.Register(context.Background(), email, "pass", "Name")
	s.Assert().Error(err, domain.ErrUserAlreadyExist)
}

func (s *AuthUsecaseSuite) TestRegister_HashError() {
	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.users.On("GetByEmail", mock.Anything, mock.Anything).Return(domain.User{}, domain.ErrUserNotFound)
	s.hasher.On("Hash", mock.Anything).Return("", errors.New("hash failed"))

	err := s.uc.Register(context.Background(), "e@example.com", "p", "n")
	s.Error(err)
}

// ====================== Login ======================

func (s *AuthUsecaseSuite) TestLogin_Success() {
	email := "login@example.com"
	password := "secret"
	user := domain.User{ID: uuid.New(), Email: email, Password: "hashedpass"}

	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.users.On("GetByEmail", mock.Anything, email).Return(user, nil)
	s.hasher.On("Compare", user.Password, password).Return(nil)
	s.jwt.On("CreateAccess", &user).Return("access.jwt", nil)
	s.jwt.On("CreateRefresh", &user).Return("refresh.jwt", "refresh-id-123", time.Now().Add(24*time.Hour), nil)
	s.refresh.On("Save", mock.Anything, "refresh-id-122", mock.AnythingOfType("time.Duration")).Return(nil)

	jwt, err := s.uc.Login(context.Background(), email, password)
	s.NoError(err)
	s.NotEmpty(jwt.Access)
	s.NotEmpty(jwt.Refresh)
}

func (s *AuthUsecaseSuite) TestLogin_UserNotFound() {
	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.users.On("GetByEmail", mock.Anything, mock.Anything).Return(domain.User{}, domain.ErrUserNotFound)

	_, err := s.uc.Login(context.Background(), "notfound@example.com", "pass")
	s.Assert().Error(err, domain.ErrUserNotFound)
}

func (s *AuthUsecaseSuite) TestLogin_InvalidCredentials() {
	user := domain.User{Password: "hashedpass"}
	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.users.On("GetByEmail", mock.Anything, mock.Anything).Return(user, nil)
	s.hasher.On("Compare", mock.Anything, mock.Anything).Return(domain.ErrInvalidCredentials)

	_, err := s.uc.Login(context.Background(), "user@example.com", "wrongpass")
	s.Assert().Error(err, domain.ErrInvalidCredentials)
}

// ====================== Refresh ======================

func (s *AuthUsecaseSuite) TestRefresh_Success() {
	refreshToken := "old.refresh.token"
	claims := &domain.RefreshClaims{UserID: uuid.New()}
	user := domain.User{ID: claims.UserID}

	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.jwt.On("ParseRefresh", refreshToken).Return(claims, nil)
	s.refresh.On("Consume", mock.Anything, claims.ID).Return(nil)
	s.users.On("GetByID", mock.Anything, claims.UserID).Return(user, nil)
	s.jwt.On("CreateAccess", &user).Return("new.access.jwt", nil)
	s.jwt.On("CreateRefresh", &user).Return("new.refresh.jwt", "new-id-789", time.Now().Add(24*time.Hour), nil)
	s.refresh.On("Save", mock.Anything, "new-id-789", mock.AnythingOfType("time.Duration")).Return(nil)

	_, err := s.uc.Refresh(context.Background(), refreshToken)
	s.NoError(err)
}

func (s *AuthUsecaseSuite) TestRefresh_ParseError() {
	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.jwt.On("ParseRefresh", mock.Anything).Return((*domain.RefreshClaims)(nil), errors.New("invalid token"))

	_, err := s.uc.Refresh(context.Background(), "bad.token")
	s.Error(err)
}

func (s *AuthUsecaseSuite) TestRefresh_ConsumeError() {
	claims := &domain.RefreshClaims{UserID: uuid.New()}
	s.uow.On("DoWithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.jwt.On("ParseRefresh", mock.Anything).Return(claims, nil)
	s.refresh.On("Consume", mock.Anything, claims.ID).Return(errors.New("consume failed"))

	_, err := s.uc.Refresh(context.Background(), "token")
	s.Error(err)
}

// ====================== Logout ======================

func (s *AuthUsecaseSuite) TestLogout_Success() {
	token := "logout.token"
	claims := &domain.RefreshClaims{}

	s.jwt.On("ParseRefresh", token).Return(claims, nil)
	s.uow.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	s.refresh.On("Delete", mock.Anything, claims.ID).Return(nil)

	err := s.uc.Logout(context.Background(), token)
	s.NoError(err)
}

func (s *AuthUsecaseSuite) TestLogout_ParseError() {
	s.jwt.On("ParseRefresh", mock.Anything).Return((*domain.RefreshClaims)(nil), errors.New("parse error"))

	err := s.uc.Logout(context.Background(), "bad.token")
	s.Error(err)
}