package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ApplicationConfig ApplicationConfig
}

type ApplicationConfig struct {
	Environment  string
	ServerPort   string
	ServerHost   string
	LogLevel     string
	JWTSecretKey string
	JWTTokenTTL  time.Duration
}

func LoadConfig(envFile ...string) (*Config, error) {
	var err error
	if len(envFile) == 0 {
		err = godotenv.Load()
	} else {
		err = godotenv.Load(envFile[0])
	}

	if err != nil {
		if !os.IsNotExist(err) && !strings.Contains(err.Error(), "no such file or directory") {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}
	jwtTokenTTL, err := time.ParseDuration(getEnv("JWT_TOKEN_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_TOKEN_TTL: %w", err)
	}

	config := &Config{
		ApplicationConfig: ApplicationConfig{
			Environment:  getEnv("ENV", "local"),
			ServerPort:   getEnv("SERVER_PORT", "8080"),
			ServerHost:   getEnv("SERVER_HOST", "0.0.0.0"),
			LogLevel:     getEnv("LOG_LEVEL", "info"),
			JWTSecretKey: getEnv("JWT_SECRET_KEY", "your-secret-key-here"),
			JWTTokenTTL:  jwtTokenTTL,
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.ApplicationConfig.JWTSecretKey) == "" {
		return fmt.Errorf("JWT_SECRET_KEY cannot be empty")
	}
	if c.ApplicationConfig.JWTTokenTTL <= 0 {
		return fmt.Errorf("JWT_TOKEN_TTL must be greater than zero")
	}
	return nil
}
