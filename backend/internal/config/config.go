package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              string
	DatabaseURL       string
	TrustProxy        bool
	ResendAPIKey      string
	ContactNotifyFrom string
	ContactNotifyTo   string
}

func Load() Config {
	loadEnvFile(".env")
	return Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://portfolio:portfolio@localhost:5433/portfolio?sslmode=disable"),
		TrustProxy:        getBool("TRUST_PROXY", false),
		ResendAPIKey:      getEnv("RESEND_API_KEY", ""),
		ContactNotifyFrom: getEnv("CONTACT_NOTIFY_FROM", "Portfolio contact form <onboarding@resend.dev>"),
		ContactNotifyTo:   getEnv("CONTACT_NOTIFY_TO", ""),
	}
}

// loadEnvFile sets variables from a KEY=VALUE file without overriding ones already set.
// A missing file is fine: in production the variables come from the environment.
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
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
