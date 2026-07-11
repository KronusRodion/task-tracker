package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/KronusRodion/task-tracker/internal/persistence"

	"github.com/google/uuid"
)

type AuthUsecase struct {
	users   UserRepository
	refresh RefreshTokenRepository
	hasher  PasswordHasher
	jwt     JWTManager
	uow     persistence.UnitOfWork
}

func New(users UserRepository, refresh RefreshTokenRepository, hasher PasswordHasher, jwt JWTManager, uow persistence.UnitOfWork) AuthUsecase {
	return AuthUsecase{
		users: users, refresh: refresh, hasher: hasher, jwt: jwt, uow: uow,
	}
}

func (u AuthUsecase) Register(
	ctx context.Context,
	email string,
	password string,
	fullName string,
) error {

	return u.uow.DoWithTx(ctx, func(ctx context.Context) error {

		user, err := u.users.GetByEmail(ctx, email)

		switch {
		case err == nil:
			return domain.ErrUserAlreadyExist

		case !errors.Is(err, domain.ErrUserNotFound):
			return err
		}

		hash, err := u.hasher.Hash(password)
		if err != nil {
			return err
		}

		user = domain.User{
			ID:        uuid.New(),
			Email:     email,
			FullName:  fullName,
			Password:  hash,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		return u.users.Create(ctx, user)
	})
}

func (u AuthUsecase) Login(
	ctx context.Context,
	email string,
	password string,
) (domain.JWT, error) {
	var jwt domain.JWT
	err := u.uow.DoWithTx(ctx, func(ctx context.Context) error {
		user, err := u.users.GetByEmail(ctx, email)
		if err != nil {
			return err
		}

		if err := u.hasher.Compare(user.Password, password); err != nil {
			return domain.ErrInvalidCredentials
		}

		access, err := u.jwt.CreateAccess(&user)
		if err != nil {
			return err
		}

		refresh, tokenID, exp, err := u.jwt.CreateRefresh(&user)
		if err != nil {
			return err
		}

		if err := u.refresh.Save(
			ctx,
			tokenID,
			time.Until(exp),
		); err != nil {
			return err
		}

		jwt = domain.JWT{
			Access:  access,
			Refresh: refresh,
		}
		return nil
	})

	return jwt, err
}

func (u AuthUsecase) Refresh(
	ctx context.Context,
	refreshToken string,
) (domain.JWT, error) {

	var jwt domain.JWT
	err := u.uow.DoWithTx(ctx, func(ctx context.Context) error {
		claims, err := u.jwt.ParseRefresh(refreshToken)
		if err != nil {
			return err
		}

		if err := u.refresh.Consume(ctx, claims.ID); err != nil {
			return err
		}

		user, err := u.users.GetByID(ctx, claims.UserID)
		if err != nil {
			return err
		}

		access, err := u.jwt.CreateAccess(&user)
		if err != nil {
			return err
		}

		refresh, tokenID, exp, err := u.jwt.CreateRefresh(&user)
		if err != nil {
			return err
		}

		if err := u.refresh.Save(
			ctx,
			tokenID,
			time.Until(exp),
		); err != nil {
			return err
		}

		jwt = domain.JWT{
			Access:  access,
			Refresh: refresh,
		}

		return nil
	})

	return jwt, err
}

func (u AuthUsecase) Logout(
	ctx context.Context,
	refreshToken string,
) error {

	claims, err := u.jwt.ParseRefresh(refreshToken)
	if err != nil {
		return err
	}

	return u.uow.Do(ctx, func(ctx context.Context) error {
		return u.refresh.Delete(ctx, claims.ID)
	})
}

func (u AuthUsecase) Authenticate(token string) (*domain.UserCtx, error) {
	claims, err := u.jwt.ParseAccess(token)
	if err != nil {
		return nil, err
	}

	userCtx := &domain.UserCtx{ID: claims.UserID}

	return userCtx, nil
}
