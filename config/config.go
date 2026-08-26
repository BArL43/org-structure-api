package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
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
	cfg := &Config{ServerPort: getEnv("SERVER_PORT", "8080")}
	if err := validatePort("SERVER_PORT", cfg.ServerPort); err != nil {
		return nil, err
	}

	var err error
	if cfg.DB.Host, err = getRequiredEnv("DB_HOST"); err != nil {
		return nil, err
	}
	cfg.DB.Port = getEnv("DB_PORT", "5432")
	if err := validatePort("DB_PORT", cfg.DB.Port); err != nil {
		return nil, err
	}
	if cfg.DB.User, err = getRequiredEnv("DB_USER"); err != nil {
		return nil, err
	}
	if cfg.DB.Password, err = getRequiredEnv("DB_PASSWORD"); err != nil {
		return nil, err
	}
	if cfg.DB.Name, err = getRequiredEnv("DB_NAME"); err != nil {
		return nil, err
	}
	cfg.DB.SSLMode = getEnv("DB_SSLMODE", "disable")
	return cfg, nil
}

func (c *Config) GetDSN() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DB.User, c.DB.Password),
		Host:   net.JoinHostPort(c.DB.Host, c.DB.Port),
		Path:   c.DB.Name,
	}
	query := u.Query()
	query.Set("sslmode", c.DB.SSLMode)
	u.RawQuery = query.Encode()
	return u.String()
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

func getRequiredEnv(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	value = strings.TrimSpace(value)
	if !exists || value == "" {
		return "", fmt.Errorf("missing required environment variable: %s", key)
	}
	return value, nil
}

func validatePort(name, value string) error {
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid %s: %q", name, value)
	}
	return nil
}
