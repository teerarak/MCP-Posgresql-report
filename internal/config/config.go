package config

import (
	"fmt"
	"os"
)

// Config holds PostgreSQL connection parameters.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Load reads connection parameters from environment variables.
func Load() *Config {
	cfg := &Config{
		Host:    getEnv("POSTGRES_HOST", "localhost"),
		Port:    getEnv("POSTGRES_PORT", "5432"),
		User:    getEnv("POSTGRES_USER", "postgres"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:  os.Getenv("POSTGRES_DB"),
		SSLMode: getEnv("POSTGRES_SSLMODE", "disable"),
	}
	return cfg
}

// DSN builds a pgx-compatible connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
