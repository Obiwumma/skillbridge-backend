// Package config manages application-wide environment variable structures.
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config aggregates all operational parameters loaded from system environment registers.
type Config struct {
	Env            string
	Port           string
	JWTSecret      string
	JWTRefreshSec  string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSslMode      string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	NatsURL        string
	Judge0URL      string
	Judge0Key      string
	ExternalAIURL  string
	ExternalAIKey  string
}

// AppConfig is the globally accessible configuration pointer.
var AppConfig *Config

// Load parses environment variables and hydrates the global AppConfig instance.
func Load() {
	_ = godotenv.Load()

	AppConfig = &Config{
		Env:            getEnv("APP_ENV", "development"),
		Port:           getEnv("APP_PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", "super-secret-key-change-in-prod"),
		JWTRefreshSec:  getEnv("JWT_REFRESH_SECRET", "refresh-secret-key-change-in-prod"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "skillbridge"),
		DBSslMode:      getEnv("DB_SSLMODE", "disable"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvInt("REDIS_DB", 0),
		NatsURL:        getEnv("NATS_URL", "nats://localhost:4222"),
		Judge0URL:      getEnv("JUDGE0_URL", "https://judge0-ce.p.rapidapi.com"),
		Judge0Key:      getEnv("JUDGE0_KEY", ""),
		ExternalAIURL:  getEnv("EXTERNAL_AI_URL", "https://api.external-ai.com"),
		ExternalAIKey:  getEnv("EXTERNAL_AI_KEY", ""),
	}
}

// getEnv resolves an environment variable by key, falling back if not found.
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvInt resolves an integer environment variable, falling back on empty or parsing issues.
func getEnvInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Printf("Invalid integer value for %s: %v, using default %d", key, err, defaultVal)
		return defaultVal
	}
	return val
}
