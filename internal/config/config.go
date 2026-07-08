package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Auth     Auth     `yaml:"auth"`
	Database DBConfig `yaml:"database"`
	Cache    Cache    `yaml:"cache"`
	Port     int      `yaml:"port"`
}

func (c *Config) Validate() error {
	err := c.Auth.Validate()
	if err != nil {
		return err
	}

	err = c.Database.Validate()
	if err != nil {
		return err
	}

	err = c.Cache.Validate()
	if err != nil {
		return err
	}

	if c.Port == 0 {
		return fmt.Errorf("0 < port < 65536")
	}

	return nil
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
