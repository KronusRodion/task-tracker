package config

import "time"

// Default возвращает конфигурацию с безопасными значениями по умолчанию.
func Default() *Config {
	return &Config{
		Port: 8080,

		Auth: Auth{
			AccessSecret:  "test-access-secret",
			RefreshSecret: "test-refresh-secret",
			Issuer:        "task-tracker",
			AccessTTL:     15 * time.Minute,
			RefreshTTL:    30 * 24 * time.Hour,
		},

		Database: DBConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "root",
			Password:     "password",
			DBName:       "task_tracker",
			SSLMode:      "disable",
			Timeout:      5 * time.Second,
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},

		Cache: Cache{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
	}
}