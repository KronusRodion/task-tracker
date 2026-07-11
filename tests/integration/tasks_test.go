package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/KronusRodion/task-tracker/tests/integration/env"
	"github.com/stretchr/testify/require"
)

func TestTaskManagement(t *testing.T) {
	client := env.GetAuthClient(t)
	var newTeamId string
	// Новый тест: создание команды, приглашение пользователя и создание задачи
	t.Run("Create team, invite user, and create task", func(t *testing.T) {
		// 1. Создаём новую команду
		teamName := "Test Team " + strconv.Itoa(env.TaskOneID) // уникальное имя
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
		require.Equal(t, http.StatusCreated, teamResp.StatusCode)

		// Парсим ответ, чтобы получить ID созданной команды
		var teamResponse struct {
			ID string `json:"id"`
		}
		err = json.NewDecoder(teamResp.Body).Decode(&teamResponse)
		require.NoError(t, err)
		teamResp.Body.Close()
		newTeamId = teamResponse.ID

		// 2. Приглашаем пользователя user2 в эту команду
		invitePayload := map[string]string{
			"user_id": env.User2ID,
			"role":    "member",
		}
		inviteJSON, _ := json.Marshal(invitePayload)
		inviteResp, err := client.Post(
			env.BaseURL+"/api/v1/teams/"+newTeamId+"/invite",
			"application/json",
			bytes.NewBuffer(inviteJSON),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, inviteResp.StatusCode)
		inviteResp.Body.Close()

		// 3. Создаём задачу в этой команде, назначая исполнителем приглашённого пользователя
		taskPayload := map[string]interface{}{
			"title":       "Task in new team",
			"team_id":     newTeamId,
			"status":      "todo",
			"assignee_id": env.User2ID,
			"priority":    domain.PriorityLow,
		}
		taskJSON, _ := json.Marshal(taskPayload)
		resp, err := client.Post(
			env.BaseURL+"/api/v1/tasks",
			"application/json",
			bytes.NewBuffer(taskJSON),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()

	})

	var newTaskID uint64
	t.Run("Create a task in existing team", func(t *testing.T) {
		taskPayload := map[string]interface{}{
			"title":       "Test Task",
			"team_id":     newTeamId,
			"status":      "todo",
			"assignee_id": env.User2ID,
			"priority":    domain.PriorityHigh,
		}
		taskJSON, _ := json.Marshal(taskPayload)
		taskResp, err := client.Post(
			env.BaseURL+"/api/v1/tasks",
			"application/json",
			bytes.NewBuffer(taskJSON),
		)
		var task struct {
			ID uint64 `json:"id"`
		}
		json.NewDecoder(taskResp.Body).Decode(&task)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, taskResp.StatusCode)
		newTaskID = task.ID
		taskResp.Body.Close()
	})

	t.Run("Filter tasks by team and status", func(t *testing.T) {
		filterResp, err := client.Get(
			env.BaseURL + "/api/v1/tasks?team_id=" + newTeamId + "&status=todo",
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, filterResp.StatusCode)
	})

	t.Run("Update a task", func(t *testing.T) {
		updatePayload := map[string]string{
			"status": "in_progress",
		}
		updateJSON, _ := json.Marshal(updatePayload)
		req, err := http.NewRequest("PATCH", env.BaseURL+"/api/v1/tasks/"+strconv.Itoa(int(newTaskID)), bytes.NewBuffer(updateJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		updateResp, err := client.Do(req)
		require.NoError(t, err)
		defer updateResp.Body.Close()
		require.Equal(t, http.StatusOK, updateResp.StatusCode)
	})

	t.Run("Fetch task history", func(t *testing.T) {
		historyResp, err := client.Get(
			env.BaseURL + "/api/v1/tasks/" + strconv.Itoa(env.TaskOneID) + "/history",
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, historyResp.StatusCode)
	})

	t.Run("Non-member attempts to create a task (should fail)", func(t *testing.T) {
		nonMemberClient := env.CreateFreshClient(t)
		taskPayload := map[string]interface{}{
			"title":       "Unauthorized Task",
			"team_id":     env.TeamAlphaID,
			"status":      "todo",
			"assignee_id": env.User2ID,
			"priority":    domain.PriorityHigh,
		}
		taskJSON, _ := json.Marshal(taskPayload)
		taskResp, err := nonMemberClient.Post(
			env.BaseURL+"/api/v1/tasks",
			"application/json",
			bytes.NewBuffer(taskJSON),
		)
		require.NoError(t, err)
		defer taskResp.Body.Close()
		require.Equal(t, http.StatusForbidden, taskResp.StatusCode)
	})
}
