package env

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// Глобальные переменные для тестов
var (
	// authClient - глобальный HTTP клиент с авторизацией
	authClient *http.Client

	// testUser - глобальный тестовый пользователь
	testUser *TestUser

	// baseURL - базовый URL для запросов
	baseURL = "http://localhost:8080"

	// initOnce - гарантирует однократную инициализацию
	initOnce sync.Once
)

// TestUser структура для хранения данных пользователя
type TestUser struct {
	Email    string
	Password string
	FullName string
	Token    string
}

// InitTestUser инициализирует тестового пользователя и создает аутентифицированный клиент
func InitTestUser(t *testing.T) {
	initOnce.Do(func() {
		// Создаем тестового пользователя
		user := CreateTestUser(t)
		testUser = user

		// Создаем HTTP клиент с автоматической установкой токена
		authClient = &http.Client{
			Transport: &authTransport{
				Token: user.Token,
			},
		}

		t.Logf("Test user initialized: %s (email: %s)", user.FullName, user.Email)
	})
}

// authTransport - кастомный транспорт для автоматической установки токена
type authTransport struct {
	Token string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Клонируем запрос, чтобы не менять оригинал
	newReq := req.Clone(req.Context())
	newReq.Header.Set("Authorization", "Bearer "+t.Token)
	newReq.Header.Set("Content-Type", "application/json")

	return http.DefaultTransport.RoundTrip(newReq)
}

func CreateFreshClient(t *testing.T) *http.Client {
	user := CreateTestUser(t)
	return &http.Client{Transport: &authTransport{user.Token}}
}

// CreateTestUser создает тестового пользователя
func CreateTestUser(t *testing.T) *TestUser {
	email := "test_" + generateRandomString(8) + "@example.com"
	password := "testpass123"
	fullName := "Test User"

	// Регистрация
	regPayload := map[string]string{
		"email":     email,
		"password":  password,
		"full_name": fullName,
	}
	regJSON, _ := json.Marshal(regPayload)

	regResp, err := http.Post(
		baseURL+"/api/v1/register",
		"application/json",
		bytes.NewBuffer(regJSON),
	)
	require.NoError(t, err)
	defer regResp.Body.Close()

	if regResp.StatusCode == http.StatusConflict {
		t.Log("User already exists, trying to login")
		return loginUser(t, email, password)
	}
	require.Equal(t, http.StatusCreated, regResp.StatusCode)

	// Логин для получения токена
	return loginUser(t, email, password)
}

func loginUser(t *testing.T, email, password string) *TestUser {
	// Создаем клиент с Jar для сохранения кук
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	client := &http.Client{
		Jar: jar,
	}

	loginPayload := map[string]string{
		"email":    email,
		"password": password,
	}
	loginJSON, _ := json.Marshal(loginPayload)

	loginResp, err := client.Post(
		baseURL+"/api/v1/login",
		"application/json",
		bytes.NewBuffer(loginJSON),
	)
	require.NoError(t, err)
	defer loginResp.Body.Close()
	require.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Парсим тело ответа (рефреш токен)
	var loginResponse struct {
		Refresh string `json:"refresh"`
	}
	err = json.NewDecoder(loginResp.Body).Decode(&loginResponse)
	require.NoError(t, err)

	// Проверяем, что access токен сохранился в куках
	url, _ := url.Parse(baseURL)
	cookies := jar.Cookies(url)
	var accessToken string
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			accessToken = cookie.Value
			break
		}
	}
	require.NotEmpty(t, accessToken, "Access token not found in cookies")

	t.Logf("Login successful. Access token in cookie: %s...", accessToken[:10])

	return &TestUser{
		Email:    email,
		Password: password,
		FullName: "Test User",
		Token:    accessToken,
	}
}

// generateRandomString генерирует случайную строку
func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// GetAuthClient возвращает глобальный аутентифицированный клиент
func GetAuthClient(t *testing.T) *http.Client {
	InitTestUser(t)
	require.NotNil(t, authClient, "Auth client not initialized")
	return authClient
}

// GetTestUser возвращает глобального тестового пользователя
func GetTestUser(t *testing.T) *TestUser {
	InitTestUser(t)
	require.NotNil(t, testUser, "Test user not initialized")
	return testUser
}
