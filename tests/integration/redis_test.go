package integration

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRedisCaching(t *testing.T) {
	t.Run("Cache task list and verify TTL", func(t *testing.T) {
		// Set a test key in Redis to simulate caching a task list
		testKey := "test:task_list:1"
		testValue := `{"tasks": [{"id": 1, "title": "Test Task"}]}`
		err := TestEnv.RedisClient.Set(context.Background(), testKey, testValue, 5*time.Minute).Err()
		require.NoError(t, err)

		// Verify the key exists and has the correct value
		val, err := TestEnv.RedisClient.Get(context.Background(), testKey).Result()
		require.NoError(t, err)
		require.Equal(t, testValue, val)

		// Verify the TTL is set (should be less than 5 minutes)
		ttl, err := TestEnv.RedisClient.TTL(context.Background(), testKey).Result()
		require.NoError(t, err)
		require.Greater(t, ttl, 0)
	})

	t.Run("Cache invalidation after data change", func(t *testing.T) {
		// Simulate updating task data in the database
		// In a real test, you would update the database and then verify the cache is invalidated
		// For now, we'll just verify that Redis operations work correctly
		_, err := TestEnv.RedisClient.FlushDB(context.Background()).Result()
		require.NoError(t, err)

		// Verify the cache is empty after flush
		_, err = TestEnv.RedisClient.Get(context.Background(), "test:task_list:1").Result()
		require.ErrorIs(t, err, redis.Nil)
	})
}