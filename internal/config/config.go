package config

import (
	"fmt"
)

type Config struct {
	Auth     Auth     `yaml:"auth"`
	Database DBConfig `yaml:"database"`
	Cache    Cache    `yaml:"cache"`
	Port     int      `yaml:"port"`
}

// Load загружает конфиг из переменных окружения
func (c *Config) Load() error {
	// Загружаем свои поля
	c.Port = getEnvInt("PORT", 8080)
	
	// Рекурсивно загружаем вложенные структуры
	if err := c.Auth.Load(); err != nil {
		return fmt.Errorf("load auth: %w", err)
	}
	if err := c.Database.Load(); err != nil {
		return fmt.Errorf("load database: %w", err)
	}
	if err := c.Cache.Load(); err != nil {
		return fmt.Errorf("load cache: %w", err)
	}
	
	return nil
}

func (c *Config) Validate() error {
	if err := c.Auth.Validate(); err != nil {
		return err
	}
	if err := c.Database.Validate(); err != nil {
		return err
	}
	if err := c.Cache.Validate(); err != nil {
		return err
	}
	if c.Port == 0 {
		return fmt.Errorf("0 < port < 65536")
	}
	return nil
}

// LoadConfig загружает конфиг из переменных окружения
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	
	if err := cfg.Load(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	
	return cfg, nil
}