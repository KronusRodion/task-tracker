package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/KronusRodion/task-tracker/tests/integration/env"
	"github.com/stretchr/testify/require"
)

func TestTeamStatsEndpoint(t *testing.T) {
	client := env.GetAuthClient(t)

	t.Run("Get team stats successfully", func(t *testing.T) {
		// Создаем команду
		teamName := "Test Team"
		teamPayload := map[string]string{
			"name": teamName,
		}
		teamJSON, _ := json.Marshal(teamPayload)
		teamResp, err := client.Post(
			env.BaseURL+"/api/v1/teams",
			"application/json",
			bytes.NewBuffer(teamJSON),
		)
		require.NoError(t, err)
		defer teamResp.Body.Close()
		require.Equal(t, http.StatusCreated, teamResp.StatusCode)

		// Получаем ID созданной команды
		var teamResponse struct {
			ID string `json:"id"`
		}
		err = json.NewDecoder(teamResp.Body).Decode(&teamResponse)
		require.NoError(t, err)
		teamID := teamResponse.ID

		// Запрашиваем статистику команды
		statsResp, err := client.Get(env.BaseURL + "/api/v1/teams/" + teamID + "/stats")
		require.NoError(t, err)
		defer statsResp.Body.Close()
		require.Equal(t, http.StatusOK, statsResp.StatusCode)

		// Десериализуем ответ
		var stats domain.TeamStats
		err = json.NewDecoder(statsResp.Body).Decode(&stats)
		require.NoError(t, err)

		// Проверяем данные статистики
		require.Equal(t, teamName, stats.TeamName)
		require.GreaterOrEqual(t, stats.MemberCount, 1) // Минимум один член (создатель команды)
	})

	t.Run("Get team stats for non-existent team", func(t *testing.T) {
		// Запрашиваем статистику несуществующей команды
		nonExistentTeamID := "00000000-0000-0000-0000-000000000000"
		statsResp, err := client.Get(env.BaseURL + "/api/v1/teams/" + nonExistentTeamID + "/stats")
		require.NoError(t, err)
		defer statsResp.Body.Close()

		// Проверяем статус ответа (ожидаем ошибку доступа или несуществующей команды)
		require.Equal(t, http.StatusForbidden, statsResp.StatusCode)
	})
}