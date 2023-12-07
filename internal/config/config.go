package config

import (
	"os"
)

type Config struct {
	Port       string
	CookieName string
	MongoUrl   string
}

func Load() Config {
	return Config{
		Port:       getEnv("PORT", "8080"),
		CookieName: getEnv("COOKIE_NAME", "cookie"),
		MongoUrl:   getEnv("MONGO_URL", "mongodb://mongo:27017"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
