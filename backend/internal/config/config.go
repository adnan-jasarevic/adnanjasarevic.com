package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://portfolio:portfolio@localhost:5433/portfolio?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
