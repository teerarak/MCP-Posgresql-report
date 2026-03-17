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
	rawDSN   string // set when DATABASE_URL is used
}

// Load reads connection parameters from environment variables.
// If DATABASE_URL is set it takes precedence over individual POSTGRES_* vars.
func Load() *Config {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return &Config{rawDSN: url, DBName: "_url"}
	}
	cfg := &Config{
		Host:     getEnv("POSTGRES_HOST", "localhost"),
		Port:     getEnv("POSTGRES_PORT", "5432"),
		User:     getEnv("POSTGRES_USER", "postgres"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
	}
	return cfg
}

// DSN builds a pgx-compatible connection string.
// Returns the raw DATABASE_URL when that env var was used.
// Omits password key when empty to avoid parse ambiguity with Unix sockets.
func (c *Config) DSN() string {
	if c.rawDSN != "" {
		return c.rawDSN
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.DBName, c.SSLMode,
	)
	if c.Password != "" {
		dsn += " password=" + c.Password
	}
	return dsn
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
