package config

import (
	"os"
)

type Config struct {
	Port                string
	MidtransServerKey   string
	MidtransEnvironment string
	JWTSecret           string
}

func Load() *Config {
	return &Config{
		Port:                getEnv("PORT", ""),
		MidtransServerKey:   getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransEnvironment: getEnv("MIDTRANS_ENV", ""),
		JWTSecret:           getEnv("JWT_SECRET", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
