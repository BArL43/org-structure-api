package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerPort string
	DB         DBConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func Load() (*Config, error) {
	cfg := &Config{}

	cfg.ServerPort = getEnv("SERVER_PORT", "8080")

	var err error
	cfg.DB.Host, err = getRequiredEnv("DB_HOST")
	if err != nil {
		return nil, err
	}

	cfg.DB.Port = getEnv("DB_PORT", "5432")

	cfg.DB.User, err = getRequiredEnv("DB_USER")
	if err != nil {
		return nil, err
	}

	cfg.DB.Password, err = getRequiredEnv("DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	cfg.DB.Name, err = getRequiredEnv("DB_NAME")
	if err != nil {
		return nil, err
	}

	cfg.DB.SSLMode = getEnv("DB_SSLMode", "disable")
	return cfg, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DB.Host,
		c.DB.Port,
		c.DB.User,
		c.DB.Password,
		c.DB.Name,
		c.DB.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getRequiredEnv(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return "", fmt.Errorf("missing required environment variable: %s", key)
	}
	return value, nil
}
