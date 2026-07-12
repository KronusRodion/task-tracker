package config

import (
	"fmt"
	"strings"
	"time"
)

type DBConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	User         string        `yaml:"user"`
	Password     string        `yaml:"password"`
	DBName       string        `yaml:"dbname"`
	SSLMode      string        `yaml:"sslmode"`
	Timeout      time.Duration `yaml:"timeout"`
	MaxOpenConns int           `yaml:"max_open_conns"`
	MaxIdleConns int           `yaml:"max_idle_conns"`
}

// Load загружает DBConfig из переменных окружения
func (c *DBConfig) Load() error {
	c.Host = getEnv("DB_HOST", "localhost")
	c.Port = getEnvInt("DB_PORT", 3306)
	c.User = getEnv("DB_USER", "root")
	c.Password = getEnv("DB_PASSWORD", "password")
	c.DBName = getEnv("DB_NAME", "task_tracker")
	c.SSLMode = getEnv("DB_SSL_MODE", "disable")
	c.Timeout = getEnvDuration("DB_TIMEOUT", 5*time.Second)
	c.MaxOpenConns = getEnvInt("DB_MAX_OPEN_CONNS", 10)
	c.MaxIdleConns = getEnvInt("DB_MAX_IDLE_CONNS", 5)
	
	return nil
}

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

func (c DBConfig) PGString() string {
	ssl := c.SSLMode
	if ssl == "" {
		ssl = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		c.Host, c.Port, c.User, c.Password, c.DBName, ssl, int(c.Timeout.Seconds()),
	)
}

func (c DBConfig) MySQLDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		c.User, c.Password, c.Host, c.Port, c.DBName,
	)
}