package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterAndLogin(t *testing.T) {
	t.Run("Successful registration and login", func(t *testing.T) {
		// Register a new user
        regPayload := map[string]string{
            "email":    "test@example.com",
            "password": "password123",
            "full_name": "Test User",
        }
		regJSON, _ := json.Marshal(regPayload)
		regResp, err := http.Post(
			"http://localhost:8080/api/v1/register",
			"application/json",
			bytes.NewBuffer(regJSON),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, regResp.StatusCode)

		// Login with the registered user
		loginPayload := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		loginJSON, _ := json.Marshal(loginPayload)
		loginResp, err := http.Post(
			"http://localhost:8080/api/v1/login",
			"application/json",
			bytes.NewBuffer(loginJSON),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, loginResp.StatusCode)
	})

	t.Run("Invalid registration (missing email)", func(t *testing.T) {
		regPayload := map[string]string{
			"password": "password123",
		}
		regJSON, _ := json.Marshal(regPayload)
		regResp, err := http.Post(
			"http://localhost:8080/api/v1/register",
			"application/json",
			bytes.NewBuffer(regJSON),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, regResp.StatusCode)
	})

	t.Run("Invalid login (wrong password)", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}
		loginJSON, _ := json.Marshal(loginPayload)
		loginResp, err := http.Post(
			"http://localhost:8080/api/v1/login",
			"application/json",
			bytes.NewBuffer(loginJSON),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, loginResp.StatusCode)
	})
}

func TestJWTAuthentication(t *testing.T) {
	t.Run("Access protected endpoint with valid JWT", func(t *testing.T) {
		// Assume we have a valid JWT token from a previous login
		// This is a placeholder for testing protected endpoints
		// In a real test, you would extract the token from the login response
		// and include it in the Authorization header.
	})
}