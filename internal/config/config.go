package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Copilot   CopilotConfig
	GitHub    GitHubConfig
	Frontend  FrontendConfig
	RateLimit RateLimitConfig
	Log       LogConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
	Host string
	Mode string
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string
	Format string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	URL string
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

// CopilotConfig holds GitHub Copilot API configuration
type CopilotConfig struct {
	APIKey string
	Model  string
	APIURL string
}

// GitHubConfig holds GitHub OAuth configuration
type GitHubConfig struct {
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	AppID         int64
	AppPrivateKey string
	WebhookSecret string
}

// FrontendConfig holds frontend URL configuration
type FrontendConfig struct {
	URL string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int
	Duration time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	expirationHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		expirationHours = 24
	}

	rateLimitRequests, err := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "100"))
	if err != nil {
		rateLimitRequests = 100
	}

	rateLimitDuration, err := time.ParseDuration(getEnv("RATE_LIMIT_DURATION", "1m"))
	if err != nil {
		rateLimitDuration = time.Minute
	}

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "default-secret-key"),
			ExpirationHours: expirationHours,
		},
		Copilot: CopilotConfig{
			APIKey: getEnv("COPILOT_API_KEY", ""),
			Model:  getEnv("COPILOT_MODEL", "gpt-4o"),
			APIURL: getEnv("COPILOT_API_URL", "https://models.inference.ai.azure.com/chat/completions"),
		},
		GitHub: GitHubConfig{
			ClientID:      getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret:  getEnv("GITHUB_CLIENT_SECRET", ""),
			RedirectURL:   getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/api/v1/auth/github/callback"),
			AppID:         getEnvAsInt("GITHUB_APP_ID", 0),
			AppPrivateKey: getEnv("GITHUB_APP_PRIVATE_KEY", ""),
			WebhookSecret: getEnv("GITHUB_WEBHOOK_SECRET", ""),
		},
		Frontend: FrontendConfig{
			URL: getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
		RateLimit: RateLimitConfig{
			Requests: rateLimitRequests,
			Duration: rateLimitDuration,
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnvAsInt(key string, defaultVal int64) int64 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultVal
	}
	return value
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return c.Server.Host + ":" + c.Server.Port
}
