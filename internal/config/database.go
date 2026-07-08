package config

import (
	"fmt"
	"strings"
	"time"
)

// DBConfig — конфигурация базы данных
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`

	Timeout      time.Duration `yaml:"timeout"`
	MaxOpenConns int           `yaml:"max_open_conns"`
	MaxIdleConns int           `yaml:"max_idle_conns"`
}

// Validate проверяет корректность конфигурации
func (c DBConfig) Validate() error {
	var errs []string

	if strings.TrimSpace(c.Host) == "" {
		errs = append(errs, "host cannot be empty")
	}
	if c.Port <= 0 || c.Port > 65535 {
		errs = append(errs, "port must be between 1 and 65535")
	}
	if strings.TrimSpace(c.User) == "" {
		errs = append(errs, "user cannot be empty")
	}
	if strings.TrimSpace(c.DBName) == "" {
		errs = append(errs, "dbname cannot be empty")
	}

	// Проверка SSLMode
	validSSL := map[string]bool{
		"disable":     true,
		"allow":       true,
		"prefer":      true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if c.SSLMode != "" && !validSSL[c.SSLMode] {
		errs = append(errs, "sslmode must be one of: disable, allow, prefer, require, verify-ca, verify-full")
	}

	if c.Timeout <= 0 {
		errs = append(errs, "timeout must be positive")
	}
	if c.MaxOpenConns <= 0 {
		errs = append(errs, "max_open_conns must be positive")
	}
	if c.MaxIdleConns < 0 || c.MaxIdleConns > c.MaxOpenConns {
		errs = append(errs, "max_idle_conns must be between 0 and max_open_conns")
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid db config: %s", strings.Join(errs, "; "))
	}

	return nil
}

// ConnString возвращает строку подключения (DSN) для PostgreSQL
func (c DBConfig) ConnString() string {
	ssl := c.SSLMode
	if ssl == "" {
		ssl = "disable"
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.DBName,
		ssl,
		int(c.Timeout.Seconds()),
	)
}
