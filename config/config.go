package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	Port                string
	MidtransServerKey   string
	MidtransEnvironment string
	JWTSecret           string
}

func Load() *Config {
	loadEnvFile()

	return &Config{
		Port:                getEnv("PORT", ""),
		MidtransServerKey:   getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransEnvironment: getEnv("MIDTRANS_ENV", ""),
		JWTSecret:           getEnv("JWT_SECRET", ""),
	}
}

func loadEnvFile() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or couldn't be loaded: %v", err)
		log.Println("Using system environment variables instead")
	} else {
		log.Println("✅ .env file loaded successfully")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
