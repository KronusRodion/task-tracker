package handler

type RegisterRequest struct {
	Email    string `json:"email" example:"user@mail.com"`
	Password string `json:"password" example:"qwerty123"`
	FullName string `json:"full_name" example:"John Doe"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"user@mail.com"`
	Password string `json:"password" example:"qwerty123"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}