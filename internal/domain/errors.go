package domain

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrTeamNotFound = errors.New("team not found")
	ErrTaskNotFound = errors.New("task not found")

	ErrAccessDenied     = errors.New("access denied")
	ErrForbidden    = errors.New("action forbidden")
	ErrAlreadyMember    = errors.New("user already in team")
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrNotTeamMember    = errors.New("user is not a team member")

	ErrInvalidStatus = errors.New("invalid task status")
	ErrInvalidRole   = errors.New("invalid team role")

	ErrTaskAlreadyClosed = errors.New("task already closed")

	ErrInvalidToken = errors.New("invalid token")
	ErrTokenRevoked = errors.New("token revoked")

	ErrInvalidCredentials = errors.New("invalid credentials")
)
