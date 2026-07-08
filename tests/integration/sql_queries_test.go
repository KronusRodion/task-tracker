package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComplexSQLQueries(t *testing.T) {
	t.Run("Teams with participant counts and done tasks", func(t *testing.T) {
		// Test JOIN query: Teams with participant counts and done tasks
		rows, err := TestEnv.DB.Query(`
			SELECT
				t.name as team_name,
				COUNT(tm.user_id) as participant_count,
				COUNT(CASE WHEN ts.status = 'done' THEN 1 END) as done_tasks_last_7_days
			FROM
				teams t
			LEFT JOIN
				team_members tm ON t.id = tm.team_id
			LEFT JOIN
				tasks ts ON t.id = ts.team_id AND ts.status = 'done' AND ts.created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
			GROUP BY
				t.id, t.name
		`)
		require.NoError(t, err)
		defer rows.Close()

		// Verify at least one row is returned
		var count int
		for rows.Next() {
			count++
		}
		require.Greater(t, count, 0, "Expected at least one team with participants and tasks")
	})

	t.Run("Top 3 users by tasks created per team", func(t *testing.T) {
		// Test recursive/query with window function: Top 3 users by tasks created per team
		rows, err := TestEnv.DB.Query(`
			WITH user_task_counts AS (
				SELECT
					u.id as user_id,
					u.email as user_email,
					t.name as team_name,
					COUNT(tasks.id) as task_count
				FROM
					users u
				JOIN
					team_members tm ON u.id = tm.user_id
				JOIN
					teams t ON tm.team_id = t.id
				LEFT JOIN
					tasks ON u.id = tasks.created_by AND t.id = tasks.team_id
				GROUP BY
					u.id, u.email, t.name
			)
			SELECT
				user_id,
				user_email,
				team_name,
				task_count
			FROM
				user_task_counts
			ORDER BY
				team_name,
				task_count DESC
			LIMIT 3
		`)
		require.NoError(t, err)
		defer rows.Close()

		// Verify at least one row is returned
		var count int
		for rows.Next() {
			count++
		}
		require.Greater(t, count, 0, "Expected at least one user-task count")
	})

	t.Run("Tasks with invalid assignee-team membership", func(t *testing.T) {
		// Test query with condition on related tables: Find tasks where assignee is not a team member
		rows, err := TestEnv.DB.Query(`
			SELECT
				tasks.id as task_id,
				tasks.title as task_title,
				tm.user_id as assignee_id,
				tm.team_id as team_id
			FROM
				tasks
			JOIN
				team_members tm ON tasks.assignee_id = tm.user_id
			LEFT JOIN
				team_members team_member ON tasks.team_id = team_member.team_id AND tasks.assignee_id = team_member.user_id
			WHERE
				team_member.user_id IS NULL
		`)
		require.NoError(t, err)
		defer rows.Close()

		// Verify no rows are returned (assuming data integrity)
		var count int
		for rows.Next() {
			count++
		}
		require.Equal(t, 0, count, "Expected no tasks with invalid assignee-team membership")
	})
}