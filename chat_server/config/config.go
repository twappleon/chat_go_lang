package config

import (
	"os"
)

type Config struct {
	MongoURI string
	DBName   string
}

func LoadConfig() *Config {
	return &Config{
		MongoURI: getEnv("MONGODB_URI", "mongodb+srv://leon456:pCGji8oL5woHZR0L@cluster0.g8y1m.mongodb.net/"),
		DBName:   getEnv("DB_NAME", "chat_app"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
