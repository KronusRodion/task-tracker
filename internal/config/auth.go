package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Auth Auth `yaml:"auth"`
}

func (c *Config) Validate() error {
	err := c.Auth.Validate()
	if err != nil {
		return err
	}

	return nil
}

type Auth struct {
	AccessSecret  string        `yaml:"access_secret"`
	RefreshSecret string        `yaml:"refresh_secret"`
	Issuer        string        `yaml:"issuer"`
	AccessTTL     time.Duration `yaml:"access_ttl"`
	RefreshTTL    time.Duration `yaml:"refresh_ttl"`
}


func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
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