package env


const (
	BaseURL = "http://localhost:8080"
)

// Константы с фиксированными UUID для тестов
const (
	// User UUIDs
	User1ID = "550e8400-e29b-41d4-a716-446655440000"
	User2ID = "550e8400-e29b-41d4-a716-446655440001"
	User3ID = "550e8400-e29b-41d4-a716-446655440002"
	
	// Team UUIDs
	TeamAlphaID = "550e8400-e29b-41d4-a716-446655440010"
	TeamBetaID  = "550e8400-e29b-41d4-a716-446655440011"
	
	// Task IDs
	TaskOneID   = 100
	TaskTwoID   = 101
	TaskThreeID = 102
)

// TestUserCredentials содержит данные тестовых пользователей из БД
var TestUsers = map[string]struct {
	ID       string
	Email    string
	Password string
	FullName string
}{
	"user1": {
		ID:       User1ID,
		Email:    "user1@example.com",
		Password: "password123", // замените на реальный пароль
		FullName: "User One",
	},
	"user2": {
		ID:       User2ID,
		Email:    "user2@example.com",
		Password: "password123",
		FullName: "User Two",
	},
	"user3": {
		ID:       User3ID,
		Email:    "user3@example.com",
		Password: "password123",
		FullName: "User Three",
	},
}

// TestTeams содержит данные тестовых команд
var TestTeams = map[string]struct {
	ID        string
	Name      string
	CreatedBy string
}{
	"alpha": {
		ID:        TeamAlphaID,
		Name:      "Team Alpha",
		CreatedBy: User1ID,
	},
	"beta": {
		ID:        TeamBetaID,
		Name:      "Team Beta",
		CreatedBy: User2ID,
	},
}

// TestTasks содержит данные тестовых задач
var TestTasks = map[int]struct {
	ID         int
	TeamID     string
	Title      string
	Status     string
	Priority   string
	CreatedBy  string
	AssigneeID string
}{
	TaskOneID: {
		ID:         TaskOneID,
		TeamID:     TeamAlphaID,
		Title:      "Task One",
		Status:     "done",
		Priority:   "medium",
		CreatedBy:  User1ID,
		AssigneeID: User2ID,
	},
	TaskTwoID: {
		ID:         TaskTwoID,
		TeamID:     TeamAlphaID,
		Title:      "Task Two",
		Status:     "in_progress",
		Priority:   "high",
		CreatedBy:  User2ID,
		AssigneeID: User1ID,
	},
	TaskThreeID: {
		ID:         TaskThreeID,
		TeamID:     TeamBetaID,
		Title:      "Task Three",
		Status:     "todo",
		Priority:   "low",
		CreatedBy:  User2ID,
		AssigneeID: User3ID,
	},
}