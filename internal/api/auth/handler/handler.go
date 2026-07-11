package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/KronusRodion/task-tracker/internal/middleware"
	"github.com/KronusRodion/task-tracker/internal/tools/handler"

	"github.com/gorilla/mux"
)

const (
	accessCookieName = "access_token"
	accessCookieTTL  = 15 * time.Minute
)

type AuthUsecase interface {
	Register(
		ctx context.Context,
		email string,
		password string,
		fullName string,
	) error

	Login(
		ctx context.Context,
		email string,
		password string,
	) (domain.JWT, error)

	Refresh(
		ctx context.Context,
		refreshToken string,
	) (domain.JWT, error)

	Logout(
		ctx context.Context,
		refreshToken string,
	) error
}

type Handler struct {
	auth AuthUsecase
}

func New(auth AuthUsecase) *Handler {
	return &Handler{
		auth: auth,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {

	r.Handle("/register", middleware.Logger(http.HandlerFunc(h.Register))).Methods(http.MethodPost)
	r.Handle("/login", http.HandlerFunc((h.Login))).Methods(http.MethodPost)
	r.Handle("/refresh", middleware.Auth(middleware.Logger(http.HandlerFunc(h.Refresh)))).Methods(http.MethodPost)
	r.Handle("/logout", middleware.Auth(middleware.Logger(http.HandlerFunc(h.Logout)))).Methods(http.MethodPost)
}

func setAccessCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(accessCookieTTL.Seconds()),
	})
}

func clearAccessCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

// Register godoc
//
//	@Summary		Регистрация пользователя
//	@Description	Создает нового пользователя
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterRequest	true	"Register request"
//	@Success		201
//	@Failure		400	{object}	handler.WriteError
//	@Failure		409	{object}	handler.WriteError
//	@Failure		500	{object}	handler.WriteError
//	@Router			/api/v1/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {

	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.WriteError(w, http.StatusBadRequest, "decoding error", err.Error())
		return
	}

	ok := domain.IsEmailValid(req.Email)
	if !ok {
		handler.WriteError(w, http.StatusBadRequest, "invalid email", "email should be valid")
		return
	}

	ok = domain.IsPasswordValid(req.Password)
	if !ok {
		handler.WriteError(w, http.StatusBadRequest, "invalid password", "password too easy")
		return
	}

	err := h.auth.Register(
		r.Context(),
		req.Email,
		req.Password,
		req.FullName,
	)

	if err != nil {

		switch {
		case errors.Is(err, domain.ErrUserAlreadyExist):
			handler.WriteError(w, http.StatusConflict, "register error", err.Error())

		default:
			log.Println("err Error: ", err)
			handler.WriteError(w, http.StatusInternalServerError, "internal service error", err.Error())
		}

		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Login godoc
//
//	@Summary		Авторизация
//	@Description	Возвращает Refresh токен и записывает Access токен в HttpOnly cookie
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest	true	"Login request"
//	@Success		200	{object}	RefreshResponse
//	@Failure		400	{object}	handler.WriteError
//	@Failure		401	{object}	handler.WriteError
//	@Failure		500	{object}	handler.WriteError
//	@Router			/api/v1/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.WriteError(w, http.StatusBadRequest, "decoding error", err.Error())
		return
	}

	tokens, err := h.auth.Login(
		r.Context(),
		req.Email,
		req.Password,
	)

	if err != nil {

		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			handler.WriteError(w, http.StatusUnauthorized, "authorize error", err.Error())

		default:
			handler.WriteError(w, http.StatusInternalServerError, "internal service error", err.Error())
		}

		return
	}

	setAccessCookie(w, tokens.Access)

	handler.WriteJSON(
		w,
		http.StatusOK,
		RefreshResponse{
			RefreshToken: tokens.Refresh,
		},
	)
}

// Refresh godoc
//
//	@Summary		Обновление JWT
//	@Description	Обновляет access и refresh токены
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RefreshRequest	true	"Refresh request"
//	@Success		200	{object}	RefreshResponse
//	@Failure		400	{object}	handler.WriteError
//	@Failure		401	{object}	handler.WriteError
//	@Failure		500	{object}	handler.WriteError
//	@Router			/api/v1/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {

	var req RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.WriteError(w, http.StatusBadRequest, "decoding error", err.Error())
		return
	}

	tokens, err := h.auth.Refresh(
		r.Context(),
		req.RefreshToken,
	)

	if err != nil {

		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			handler.WriteError(w, http.StatusUnauthorized, "invalid token", err.Error())

		case errors.Is(err, domain.ErrTokenRevoked):
			handler.WriteError(w, http.StatusUnauthorized, "refresh token revoked", err.Error())

		default:
			handler.WriteError(w, http.StatusInternalServerError, "internal service error", err.Error())
		}

		return
	}

	setAccessCookie(w, tokens.Access)

	handler.WriteJSON(
		w,
		http.StatusOK,
		RefreshResponse{
			RefreshToken: tokens.Refresh,
		},
	)
}

// Logout godoc
//
//	@Summary		Выход из системы
//	@Description	Инвалидирует refresh токен
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LogoutRequest	true	"Logout request"
//	@Success		204
//	@Failure		400	{object}	handler.WriteError
//	@Failure		401	{object}	handler.WriteError
//	@Failure		500	{object}	handler.WriteError
//	@Router			/api/v1/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {

	var req LogoutRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.WriteError(w, http.StatusBadRequest, "decoding error", err.Error())
		return
	}

	err := h.auth.Logout(
		r.Context(),
		req.RefreshToken,
	)

	if err != nil {

		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			handler.WriteError(w, http.StatusUnauthorized, "invalid token", err.Error())

		case errors.Is(err, domain.ErrTokenRevoked):
			handler.WriteError(w, http.StatusUnauthorized, "refresh token revoked", err.Error())

		default:
			handler.WriteError(w, http.StatusInternalServerError, "internal service error", err.Error())
		}

		return
	}

	clearAccessCookie(w)

	w.WriteHeader(http.StatusNoContent)
}
