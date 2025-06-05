package config

import (
	"os"
)

type Config struct {
	Port              string
	DatabaseURL       string
	AuthServiceURL    string
	UserServiceURL    string
	MidtransServerKey string
	MidtransClientKey string
	MidtransBaseURL   string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://user:password@localhost/payment_db?sslmode=disable"),
		AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8082"),
		MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", "your-midtrans-server-key"),
		MidtransClientKey: getEnv("MIDTRANS_CLIENT_KEY", "your-midtrans-client-key"),
		MidtransBaseURL:   getEnv("MIDTRANS_BASE_URL", "https://api.sandbox.midtrans.com/v2"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
