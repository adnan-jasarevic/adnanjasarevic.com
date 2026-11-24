package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	DatabaseURL        string
	ClientIPHeader     string
	ResendAPIKey       string
	ContactNotifyFrom  string
	ContactNotifyTo    string
	AllowedOrigins     []string
	GithubClientID     string
	GithubClientSecret string
	GithubRedirectURL  string
	AdminGithubID      int64
	AdminURL           string
}

func Load() Config {
	loadEnvFile(".env")
	return Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://portfolio:portfolio@localhost:5433/portfolio?sslmode=disable"),
		ClientIPHeader:     getEnv("CLIENT_IP_HEADER", ""),
		ResendAPIKey:       getEnv("RESEND_API_KEY", ""),
		ContactNotifyFrom:  getEnv("CONTACT_NOTIFY_FROM", "Portfolio contact form <onboarding@resend.dev>"),
		ContactNotifyTo:    getEnv("CONTACT_NOTIFY_TO", ""),
		AllowedOrigins:     getList("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
		GithubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GithubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GithubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/auth/github/callback"),
		AdminGithubID:      getInt64("ADMIN_GITHUB_ID", 202077371),
		AdminURL:           getEnv("ADMIN_URL", "http://localhost:5173/admin"),
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

func getList(key, fallback string) []string {
	var items []string
	for _, item := range strings.Split(getEnv(key, fallback), ",") {
		item = strings.TrimSuffix(strings.TrimSpace(item), "/")
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}

func getInt64(key string, fallback int64) int64 {
	v, err := strconv.ParseInt(getEnv(key, ""), 10, 64)
	if err != nil {
		return fallback
	}
	return v
}
