package config

import (
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OpenCode OpenCodeConfig
	GitHub   GitHubConfig
}

type ServerConfig struct {
	Host string
	Port string
	Dir  string
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	Secret     string
	Expiry     time.Duration
	CookieName string
}

type OpenCodeConfig struct {
	BinaryPath string
}

type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("HOST", "0.0.0.0"),
			Port: getEnv("PORT", "8080"),
			Dir:  getEnv("SERVE_DIR", "./web/dist"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/opencode_webchat?sslmode=disable"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "change-me-in-production"),
			Expiry:     duration(getEnv("JWT_EXPIRY", "24h")),
			CookieName: getEnv("JWT_COOKIE_NAME", "owc_token"),
		},
		OpenCode: OpenCodeConfig{
			BinaryPath: getEnv("OPENCODE_BINARY", "opencode"),
		},
		GitHub: GitHubConfig{
			ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
			RedirectURI:  getEnv("GITHUB_REDIRECT_URI", "http://localhost:8080/api/auth/github/callback"),
		},
	}
}

func getEnv(key, default_ string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return default_
}

func duration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

func (c *Config) Address() string {
	return c.Server.Host + ":" + c.Server.Port
}