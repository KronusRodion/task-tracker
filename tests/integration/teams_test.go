package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/KronusRodion/task-tracker/tests/integration/env"
	"github.com/stretchr/testify/require"
)

func TestTeamManagement(t *testing.T) {
	client := env.GetAuthClient(t)
	var newTeamID string
	t.Run("Create a team and list teams", func(t *testing.T) {
		teamPayload := map[string]string{
			"name": "Test Team",
		}
		teamJSON, _ := json.Marshal(teamPayload)
		teamResp, err := client.Post(
			env.BaseURL+"/api/v1/teams",
			"application/json",
			bytes.NewBuffer(teamJSON),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, teamResp.StatusCode)
		var teamResponse struct {
			ID string `json:"id"`
		}
		json.NewDecoder(teamResp.Body).Decode(&teamResponse)
		newTeamID = teamResponse.ID
	})

	t.Run("List teams where user is a member", func(t *testing.T) {
		listResp, err := client.Get(
			env.BaseURL + "/api/v1/teams",
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, listResp.StatusCode)
	})

	t.Run("Invite user to team (as owner)", func(t *testing.T) {
		invitePayload := map[string]string{
			"user_id": env.User2ID,
			"role":    "member",
		}
		inviteJSON, _ := json.Marshal(invitePayload)
		inviteResp, err := client.Post(
			env.BaseURL+"/api/v1/teams/"+newTeamID+"/invite",
			"application/json",
			bytes.NewBuffer(inviteJSON),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, inviteResp.StatusCode)
	})

	t.Run("Non-owner attempts to invite user (should fail)", func(t *testing.T) {
		// Создаём клиент от имени пользователя, не являющегося владельцем TeamAlpha
		nonOwnerClient := env.CreateFreshClient(t)
		invitePayload := map[string]string{
			"user_id": env.User2ID,
			"role":    "member",
		}
		inviteJSON, _ := json.Marshal(invitePayload)
		inviteResp, err := nonOwnerClient.Post(
			env.BaseURL+"/api/v1/teams/"+env.TeamAlphaID+"/invite",
			"application/json",
			bytes.NewBuffer(inviteJSON),
		)
		require.NoError(t, err)
		defer inviteResp.Body.Close()
		require.Equal(t, http.StatusForbidden, inviteResp.StatusCode)
	})
}
