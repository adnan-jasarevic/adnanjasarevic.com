package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DatabaseURL string
	TrustProxy  bool
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://portfolio:portfolio@localhost:5433/portfolio?sslmode=disable"),
		TrustProxy:  getBool("TRUST_PROXY", false),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	v, err := strconv.ParseBool(getEnv(key, ""))
	if err != nil {
		return fallback
	}
	return v
}
