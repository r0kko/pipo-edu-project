package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env               string
	HTTPAddr          string
	DBDSN             string
	JWTSecret         string
	JWTRefreshSecret  string
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
	MigrateOnStart    bool
	CORSOrigins       []string
	BootstrapEmail    string
	BootstrapPassword string
	BootstrapName     string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Env:               getEnv("APP_ENV", "dev"),
		HTTPAddr:          getEnv("HTTP_ADDR", ":8080"),
		DBDSN:             getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5432/pipo?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-access"),
		JWTRefreshSecret:  getEnv("JWT_REFRESH_SECRET", "change-me-refresh"),
		AccessTTL:         getEnvDuration("ACCESS_TTL", 15*time.Minute),
		RefreshTTL:        getEnvDuration("REFRESH_TTL", 7*24*time.Hour),
		MigrateOnStart:    getEnvBool("MIGRATE_ON_START", true),
		CORSOrigins:       getEnvCSV("CORS_ORIGINS", ""),
		BootstrapEmail:    getEnv("BOOTSTRAP_ADMIN_EMAIL", ""),
		BootstrapPassword: getEnv("BOOTSTRAP_ADMIN_PASSWORD", ""),
		BootstrapName:     getEnv("BOOTSTRAP_ADMIN_NAME", "Администратор"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvCSV(key, fallback string) []string {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}
	if value == "" {
		return nil
	}
	items := []string{}
	start := 0
	for i := 0; i <= len(value); i++ {
		if i == len(value) || value[i] == ',' {
			item := value[start:i]
			if item != "" {
				items = append(items, item)
			}
			start = i + 1
		}
	}
	return items
}
