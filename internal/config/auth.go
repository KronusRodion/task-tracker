package config

import (
	"fmt"
	"time"
)

type Auth struct {
	AccessSecret  string        `yaml:"access_secret"`
	RefreshSecret string        `yaml:"refresh_secret"`
	Issuer        string        `yaml:"issuer"`
	AccessTTL     time.Duration `yaml:"access_ttl"`
	RefreshTTL    time.Duration `yaml:"refresh_ttl"`
}

// Load загружает Auth из переменных окружения
func (a *Auth) Load() error {
	a.AccessSecret = getEnv("AUTH_ACCESS_SECRET", "test-access-secret")
	a.RefreshSecret = getEnv("AUTH_REFRESH_SECRET", "test-refresh-secret")
	a.Issuer = getEnv("AUTH_ISSUER", "task-tracker")
	a.AccessTTL = getEnvDuration("AUTH_ACCESS_TTL", 15*time.Minute)
	a.RefreshTTL = getEnvDuration("AUTH_REFRESH_TTL", 30*24*time.Hour)
	
	return nil
}

func (a *Auth) Validate() error {
	if a.AccessSecret == "" {
		return fmt.Errorf("auth.access_secret is required")
	}
	if a.RefreshSecret == "" {
		return fmt.Errorf("auth.refresh_secret is required")
	}
	if a.Issuer == "" {
		return fmt.Errorf("auth.issuer is required")
	}
	if a.AccessTTL <= 0 {
		return fmt.Errorf("auth.access_ttl must be > 0")
	}
	if a.RefreshTTL <= 0 {
		return fmt.Errorf("auth.refresh_ttl must be > 0")
	}
	return nil
}