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
		MongoURI: getEnv("MONGODB_URI", "mongodb://localhost:27017"),
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
