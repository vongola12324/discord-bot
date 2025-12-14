package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	DiscordToken string
}

var instance *Config

// Load initializes and returns the application configuration
func Load() (*Config, error) {
	if instance != nil {
		return instance, nil
	}

	// Load .env file if exists (ignore error if not found)
	_ = godotenv.Load()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN environment variable is required")
	}

	instance = &Config{
		DiscordToken: token,
	}

	return instance, nil
}

// Get returns the singleton config instance
func Get() *Config {
	return instance
}
