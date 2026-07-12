package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/KronusRodion/task-tracker/internal/middleware"
	"github.com/KronusRodion/task-tracker/internal/tools/handler"
)

type TeamsUsecase interface {
	CreateTeam(
		ctx context.Context,
		name string,
		ownerID uuid.UUID,
	) (domain.Team, error)

	GetUserTeams(
		ctx context.Context,
		userID uuid.UUID,
	) ([]domain.Team, error)

	InviteUser(
		ctx context.Context,
		teamID uuid.UUID,
		userID uuid.UUID,
		invitedBy uuid.UUID,
		role domain.TeamRole,
	) error

	GetTeamStats(
		ctx context.Context,
		teamID uuid.UUID,
		userID uuid.UUID,
	) (domain.TeamStats, error)
}

type Handler struct {
	teams TeamsUsecase
}

func New(teams TeamsUsecase) *Handler {
	return &Handler{
		teams: teams,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	api := r.PathPrefix("/teams").Subrouter()

	api.Use(middleware.Auth)
	api.Use(middleware.Logger)
	api.HandleFunc("", h.CreateTeam).Methods(http.MethodPost)
	api.HandleFunc("", h.ListTeams).Methods(http.MethodGet)
	api.HandleFunc("/{id}/stats", h.GetTeamStats).Methods(http.MethodGet)
	api.HandleFunc("/{id}/invite", h.InviteUser).Methods(http.MethodPost)
}

// CreateTeam godoc
//
//	@Summary		Создание команды
//	@Description	Создает команду и назначает пользователя owner
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateTeamRequest	true	"Create team request"
//	@Success		201	{object}	TeamResponse
//	@Failure		400	{object}	handler.WriteError
//	@Failure		401	{object}	handler.WriteError
//	@Failure		500	{object}	handler.WriteError
//	@Router			/api/v1/teams [post]
func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {

	var req CreateTeamRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.WriteError(w, http.StatusBadRequest, "decode error", err.Error())
		return
	}

	userID, err := domain.UserIDFromContext(r.Context())
	if err != nil {
		handler.WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	team, err := h.teams.CreateTeam(r.Context(), req.Name, userID)
	if err != nil {
		handler.WriteError(w, http.StatusInternalServerError, "create team error", err.Error())
		return
	}

	handler.WriteJSON(w, http.StatusCreated, TeamResponse{
		ID:   team.ID,
		Name: team.Name,
	})
}

// ListTeams godoc
//
//	@Summary		Список команд пользователя
//	@Description	Возвращает команды, в которых состоит пользователь
//	@Tags			teams
//	@Produce		json
//	@Success		200	{object}	ListTeamsResponse
//	@Failure		401	{object}	handler.WriteError
//	@Failure		500	{object}	handler.WriteError
//	@Router			/api/v1/teams [get]
func (h *Handler) ListTeams(w http.ResponseWriter, r *http.Request) {

	userID, err := domain.UserIDFromContext(r.Context())
	if err != nil {
		handler.WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	teams, err := h.teams.GetUserTeams(r.Context(), userID)
	if err != nil {
		handler.WriteError(w, http.StatusInternalServerError, "get teams error", err.Error())
		return
	}

	resp := ListTeamsResponse{
		Teams: make([]TeamResponse, 0, len(teams)),
	}

	for _, t := range teams {
		resp.Teams = append(resp.Teams, TeamResponse{
			ID:   t.ID,
			Name: t.Name,
		})
	}

	handler.WriteJSON(w, http.StatusOK, resp)
}

// InviteUser godoc
//
//	@Summary		Приглашение пользователя в команду
//	@Description	Добавляет пользователя в команду (owner/admin only)
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Team ID"
//	@Param			request	body		InviteUserRequest	true	"Invite request"
//	@Success		204
//	@Failure		400	{object}	handler.WriteError
//	@Failure		401	{object}	handler.WriteError
//	@Failure		403	{object}	handler.WriteError
//	@Failure		500	{object}	handler.WriteError
//	@Router			/api/v1/teams/{id}/invite [post]
func (h *Handler) InviteUser(w http.ResponseWriter, r *http.Request) {

	var req InviteUserRequest

	userID, err := domain.UserIDFromContext(r.Context())
	if err != nil {
		handler.WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.WriteError(w, http.StatusBadRequest, "decode error", err.Error())
		return
	}

	teamIDStr := mux.Vars(r)["id"]
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		handler.WriteError(w, http.StatusBadRequest, "invalid team id", err.Error())
		return
	}

	err = h.teams.InviteUser(
		r.Context(),
		teamID,
		req.UserID,
		userID,
		req.Role,
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrNotTeamMember):
			handler.WriteError(w, http.StatusForbidden, "forbidden", err.Error())

		case errors.Is(err, domain.ErrUserNotFound):
			handler.WriteError(w, http.StatusNotFound, "user not found", err.Error())

		default:
			handler.WriteError(w, http.StatusInternalServerError, "invite error", err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAllTeamsStats godoc
//
//	@Summary		Получение статистики по всем командам
//	@Description	Возвращает статистику по всем командам (только для админов)
//	@Tags			teams
//	@Produce		json
//	@Success		200	{array}	domain.TeamStats
//	@Failure		401	{object}	handler.WriteError
//	@Failure		403	{object}	handler.WriteError
//	@Failure		500	{object}	handler.WriteError
//	@Router			/api/v1/teams/stats [get]
func (h *Handler) GetTeamStats(w http.ResponseWriter, r *http.Request) {
	userID, err := domain.UserIDFromContext(r.Context())
	if err != nil {
		handler.WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	teamIDStr := mux.Vars(r)["id"]
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		handler.WriteError(w, http.StatusBadRequest, "invalid team id", err.Error())
		return
	}

	stats, err := h.teams.GetTeamStats(r.Context(), teamID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrTeamNotFound):
			handler.WriteError(w, http.StatusForbidden, "forbidden", err.Error())
		default:
			handler.WriteError(w, http.StatusInternalServerError, "get team stats error", err.Error())
		}
		return
	}

	handler.WriteJSON(w, http.StatusOK, stats)
}
