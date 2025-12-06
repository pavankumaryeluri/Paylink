package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DBUser              string
	DBPassword          string
	DBHost              string
	DBPort              string
	DBName              string
	RedisHost           string
	RedisPort           string
	MidtransServerKey   string
	XenditAPIKey        string
	StripeSecretKey     string
	StripeWebhookSecret string
}

func LoadConfig() *Config {
	_ = godotenv.Load() // Ignore error if .env not found (system env vars)

	return &Config{
		Port:                getEnv("PORT", "8080"),
		DBUser:              getEnv("DB_USER", "paylink"),
		DBPassword:          getEnv("DB_PASSWORD", "secret"),
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "5432"),
		DBName:              getEnv("DB_NAME", "paylink"),
		RedisHost:           getEnv("REDIS_HOST", "localhost"),
		RedisPort:           getEnv("REDIS_PORT", "6379"),
		MidtransServerKey:   getEnv("MIDTRANS_SERVER_KEY", ""),
		XenditAPIKey:        getEnv("XENDIT_API_KEY", ""),
		StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
