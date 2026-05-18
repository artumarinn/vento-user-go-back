package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	ServerPort        string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	JWTSecret         string
	Env               string
	CoreInternalToken string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults and system environment variables")
	}

	return &Config{
		ServerPort:        getEnv("SERVER_PORT", "8082"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5433"),
		DBUser:            getEnv("DB_USER", "vento"),
		DBPassword:        getEnv("DB_PASSWORD", "vento_secret"),
		DBName:            getEnv("DB_NAME", "vento_db"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		JWTSecret:         getEnv("JWT_SECRET", "vento-dev-secret-change-in-production"),
		Env:               getEnv("ENV", "development"),
		CoreInternalToken: getEnv("CORE_INTERNAL_TOKEN", ""),
	}
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
