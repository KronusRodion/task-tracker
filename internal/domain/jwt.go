package domain

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWT struct {
	Access  string
	Refresh string
}

type AccessClaims struct {
	UserID uuid.UUID
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID uuid.UUID
	jwt.RegisteredClaims
}
