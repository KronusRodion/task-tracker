package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/KronusRodion/task-tracker/tests/integration/env"
	"github.com/stretchr/testify/require"
)

func TestTaskManagement(t *testing.T) {
	client := env.GetAuthClient(t)

	t.Run("Create a task", func(t *testing.T) {
		taskPayload := map[string]interface{}{
			"title":       "Test Task",
			"team_id":     env.TeamAlphaID, // UUID команды
			"status":      "todo",
			"assignee_id": env.User2ID, // UUID пользователя
		}
		taskJSON, _ := json.Marshal(taskPayload)
		taskResp, err := client.Post(
			env.BaseURL+"/api/v1/tasks",
			"application/json",
			bytes.NewBuffer(taskJSON),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, taskResp.StatusCode)
	})

	t.Run("Filter tasks by team and status", func(t *testing.T) {
		filterResp, err := client.Get(
			env.BaseURL + "/api/v1/tasks?team_id=" + env.TeamAlphaID + "&status=todo",
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, filterResp.StatusCode)
	})

	t.Run("Update a task", func(t *testing.T) {
		updatePayload := map[string]string{
			"status": "in_progress",
		}
		updateJSON, _ := json.Marshal(updatePayload)
		req, err := http.NewRequest("PATCH", env.BaseURL+"/api/v1/tasks/"+strconv.Itoa(env.TaskOneID), bytes.NewBuffer(updateJSON))
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
		// Создаём клиент от имени пользователя, не состоящего в команде TeamAlpha
		nonMemberClient := env.CreateFreshClient(t) // создаёт нового пользователя
		taskPayload := map[string]interface{}{
			"title":       "Unauthorized Task",
			"team_id":     env.TeamAlphaID,
			"status":      "todo",
			"assignee_id": env.User2ID,
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
