package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ApplicationConfig ApplicationConfig
	DBConfig          DBConfig
}

type ApplicationConfig struct {
	Environment string
	ServerPort  string
	ServerHost  string
	LogLevel    string
}

type DBConfig struct {
	DSN string // full NeonDB connection string from env
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

	config := &Config{
		ApplicationConfig: ApplicationConfig{
			Environment: getEnv("ENV", "local"),
			ServerPort:  getEnv("SERVER_PORT", "8080"),
			ServerHost:  getEnv("SERVER_HOST", "0.0.0.0"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
		DBConfig: DBConfig{
			DSN: getEnv("DATABASE_URL", ""),
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
	if c.DBConfig.DSN == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}
