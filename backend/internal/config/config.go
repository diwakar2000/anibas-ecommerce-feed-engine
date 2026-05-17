package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	APIPort        string
	FrontendOrigin string
	PostgresDB     string
	PostgresUser   string
	PostgresPass   string
	PostgresHost   string
	PostgresPort   string
	PostgresSSL    string
}

func Load() Config {
	return Config{
		APIPort:        env("API_PORT", "8080"),
		FrontendOrigin: env("FRONTEND_ORIGIN", "http://localhost:5173"),
		PostgresDB:     env("POSTGRES_DB", "anibas_feed"),
		PostgresUser:   env("POSTGRES_USER", "anibas"),
		PostgresPass:   env("POSTGRES_PASSWORD", "anibas_password"),
		PostgresHost:   env("POSTGRES_HOST", "localhost"),
		PostgresPort:   env("POSTGRES_PORT", "5432"),
		PostgresSSL:    env("POSTGRES_SSLMODE", "disable"),
	}
}

func (c Config) DatabaseURL() string {
	user := url.UserPassword(c.PostgresUser, c.PostgresPass)
	return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s", user.String(), c.PostgresHost, c.PostgresPort, c.PostgresDB, c.PostgresSSL)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
