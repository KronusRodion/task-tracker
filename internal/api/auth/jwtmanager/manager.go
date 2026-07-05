package jwtmanager

import (
	"fmt"
	"time"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Manager struct {
	accessSecret  []byte
	refreshSecret []byte

	issuer string

	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(
	accessSecret string,
	refreshSecret string,
	issuer string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *Manager {

	return &Manager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		issuer:        issuer,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (m *Manager) CreateAccess(
	user *domain.User,
) (string, error) {

	now := time.Now()

	claims := domain.AccessClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: uuid.NewString(),

			Issuer: m.issuer,

			IssuedAt: jwt.NewNumericDate(now),

			ExpiresAt: jwt.NewNumericDate(
				now.Add(m.accessTTL),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS512,
		claims,
	)

	return token.SignedString(m.accessSecret)
}

func (m *Manager) CreateRefresh(
	user *domain.User,
) (
	token string,
	tokenID string,
	exp time.Time,
	err error,
) {

	now := time.Now()

	tokenID = uuid.NewString()

	exp = now.Add(m.refreshTTL)

	claims := domain.RefreshClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:     tokenID,
			Issuer: m.issuer,

			IssuedAt: jwt.NewNumericDate(now),

			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	j := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	token, err = j.SignedString(m.refreshSecret)

	return
}

func (m *Manager) ParseAccess(
	token string,
) (*domain.AccessClaims, error) {

	claims := new(domain.AccessClaims)

	parsed, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (interface{}, error) {

			if t.Method != jwt.SigningMethodHS512 {
				return nil, fmt.Errorf("unexpected signing method")
			}

			return m.accessSecret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	if !parsed.Valid {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

func (m *Manager) ParseRefresh(
	token string,
) (*domain.RefreshClaims, error) {

	claims := new(domain.RefreshClaims)

	parsed, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (interface{}, error) {

			if t.Method != jwt.SigningMethodHS512 {
				return nil, fmt.Errorf("unexpected signing method")
			}

			return m.refreshSecret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	if !parsed.Valid {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}
