package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Cache struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`
}

// Load загружает Cache из переменных окружения
func (c *Cache) Load() error {
	c.Host = getEnv("REDIS_HOST", "localhost")
	c.Port = getEnvInt("REDIS_PORT", 6379)
	c.Password = getEnv("REDIS_PASSWORD", "")
	c.DB = getEnvInt("REDIS_DB", 0)
	
	return nil
}

func (c *Cache) Validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return errors.New("redis host cannot be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d (must be between 1 and 65535)", c.Port)
	}
	if c.DB < 0 || c.DB > 15 {
		return fmt.Errorf("invalid redis DB index: %d (must be between 0 and 15)", c.DB)
	}
	return nil
}

func (c *Cache) Address() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}