package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL    string
	RedisAddr      string
	RedisPass      string
	RedisDB        int
	JWTSecret      string
	JWTTTL         time.Duration
	APIPort        string
	CORSOrigins    []string
	ScanTimeout    time.Duration
	SeedDemo       bool
	TrustedProxies []string
	MaxScanRetries int
}

func Load() Config {
	return Config{
		DatabaseURL:    env("DATABASE_URL", "postgres://quantumfield:quantumfield_dev@localhost:5432/quantumfield?sslmode=disable"),
		RedisAddr:      env("REDIS_ADDR", "localhost:6379"),
		RedisPass:      os.Getenv("REDIS_PASSWORD"),
		RedisDB:        envInt("REDIS_DB", 0),
		JWTSecret:      env("JWT_SECRET", "development-secret-change-me-please"),
		JWTTTL:         time.Duration(envInt("JWT_TTL_HOURS", 24)) * time.Hour,
		APIPort:        env("API_PORT", "8080"),
		CORSOrigins:    splitCSV(env("CORS_ORIGINS", "http://localhost:5173")),
		ScanTimeout:    time.Duration(envInt("SCAN_TIMEOUT_SECONDS", 15)) * time.Second,
		SeedDemo:       envBool("SEED_DEMO", false),
		TrustedProxies: splitCSV(os.Getenv("TRUSTED_PROXIES")),
		MaxScanRetries: envInt("MAX_SCAN_RETRIES", 3),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			result = append(result, item)
		}
	}
	return result
}
