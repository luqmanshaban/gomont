package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT       int
	DB_PORT    int
	DB_NAME    string
	DB_HOST    string
	DB_PASS    string
	DB_USER    string
	EMAIL_USER string
	EMAIL_PASS string
	EMAIL_HOST string
	EMAIL_PORT int
	JWT_SECRET string
}

func LoadEnv() *Config {
	_ = godotenv.Load()

	return &Config{
		PORT:       getEnvAsInt("PORT", 8000),
		DB_PORT:    getEnvAsInt("DB_PORT", 5432),
		DB_NAME:    getEnv("DB_NAME", ""),
		DB_HOST:    getEnv("DB_HOST", ""),
		DB_PASS:    getEnv("DB_PASS", ""),
		DB_USER:    getEnv("DB_USER", ""),
		EMAIL_USER: getEnv("EMAIL_USER", ""),
		EMAIL_PASS: getEnv("EMAIL_PASS", ""),
		EMAIL_HOST: getEnv("EMAIL_HOST", ""),
		EMAIL_PORT: getEnvAsInt("EMAIL_PORT", 587),
		JWT_SECRET: getEnv("JWT_SECRET",""),
	}
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err != nil {
		return val
	}
	return defaultVal
}
