package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppPort     string
	AppEnv      string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	JWTSecret   string
	JWTExpHours int
	VTAPIKey    string
}

func Load() Config {
	return Config{
		AppPort:     getEnv("APP_PORT", "8080"),
		AppEnv:      getEnv("APP_ENV", "development"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "3306"),
		DBUser:      getEnv("DB_USER", "ticketguard"),
		DBPassword:  getEnv("DB_PASSWORD", "ticketguard123"),
		DBName:      getEnv("DB_NAME", "ticketguard_db"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpHours: getEnvAsInt("JWT_EXP_HOURS", 8),
		VTAPIKey:    getEnv("VT_API_KEY", ""),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func (c Config) JWTDuration() time.Duration {
	return time.Duration(c.JWTExpHours) * time.Hour
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
