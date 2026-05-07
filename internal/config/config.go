package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL          string
	JWTSecret            string
	Port                 string
	AdminEmail           string
	CORSAllowedOrigin    string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	BcryptCost           int
	RateLimitAttempts    int
	RateLimitWindow      time.Duration
}

func New() (*Config, error) {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	accessDuration, err := parseDurationEnv("ACCESS_TOKEN_DURATION", 15*time.Minute)
	if err != nil {
		return nil, err
	}

	refreshDuration, err := parseDurationEnv("REFRESH_TOKEN_DURATION", 168*time.Hour)
	if err != nil {
		return nil, err
	}

	rateLimitWindow, err := parseDurationEnv("RATE_LIMIT_WINDOW", time.Minute)
	if err != nil {
		return nil, err
	}

	bcryptCost, err := parseIntEnv("BCRYPT_COST", 10)
	if err != nil {
		return nil, err
	}

	rateLimitAttempts, err := parseIntEnv("RATE_LIMIT_ATTEMPTS", 5)
	if err != nil {
		return nil, err
	}

	return &Config{
		Port: getEnv("PORT", "8080"),

		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,

		AdminEmail:        os.Getenv("ADMIN_EMAIL"),
		CORSAllowedOrigin: getEnv("CORS_ALLOWED_ORIGIN", "*"),

		AccessTokenDuration:  accessDuration,
		RefreshTokenDuration: refreshDuration,

		BcryptCost:        bcryptCost,
		RateLimitAttempts: rateLimitAttempts,
		RateLimitWindow:   rateLimitWindow,
	}, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func parseDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s has invalid duration value: %w", key, err)
	}

	return parsed, nil
}

func parseIntEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s has invalid integer value: %w", key, err)
	}

	return parsed, nil
}
