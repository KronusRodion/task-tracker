package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTaskManagement(t *testing.T) {
	t.Run("Create a task", func(t *testing.T) {
		// Create a task (requires authentication and team membership)
		// This is a placeholder; in a real test, you would include a valid JWT token
		taskPayload := map[string]interface{}{
			"title":     "Test Task",
			"team_id":   1,
			"status":    "todo",
			"assignee_id": 2,
		}
		taskJSON, _ := json.Marshal(taskPayload)
		taskResp, err := http.Post(
			"http://localhost:8080/api/v1/tasks",
			"application/json",
			bytes.NewBuffer(taskJSON),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, taskResp.StatusCode)
	})

	t.Run("Filter tasks by team and status", func(t *testing.T) {
		// Filter tasks (requires authentication)
		// This is a placeholder; in a real test, you would include a valid JWT token
		filterResp, err := http.Get(
			"http://localhost:8080/api/v1/tasks?team_id=1&status=todo",
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, filterResp.StatusCode)
	})

	t.Run("Update a task", func(t *testing.T) {
		// Update a task (requires authentication and permission)
		// This is a placeholder; in a real test, you would include a valid JWT token
		updatePayload := map[string]string{
			"status": "in_progress",
		}
		updateJSON, _ := json.Marshal(updatePayload)
		req, err := http.NewRequest("PUT", "http://localhost:8080/api/v1/tasks/1", bytes.NewBuffer(updateJSON))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		updateResp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer updateResp.Body.Close()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, updateResp.StatusCode)
	})

	t.Run("Fetch task history", func(t *testing.T) {
		// Fetch task history (requires authentication)
		// This is a placeholder; in a real test, you would include a valid JWT token
		historyResp, err := http.Get(
			"http://localhost:8080/api/v1/tasks/1/history",
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, historyResp.StatusCode)
	})

	t.Run("Non-member attempts to create a task (should fail)", func(t *testing.T) {
		// Non-member attempts to create a task (should return 403 Forbidden)
		// This is a placeholder; in a real test, you would include an invalid JWT token
		taskPayload := map[string]interface{}{
			"title":     "Unauthorized Task",
			"team_id":   1,
			"status":    "todo",
			"assignee_id": 2,
		}
		taskJSON, _ := json.Marshal(taskPayload)
		taskResp, err := http.Post(
			"http://localhost:8080/api/v1/tasks",
			"application/json",
			bytes.NewBuffer(taskJSON),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, taskResp.StatusCode)
	})
}