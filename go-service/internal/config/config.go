package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	Port            string
	NodeAPIURL      string
	NodeAPIUsername string
	NodeAPIPassword string
}

// Load reads configuration from the environment, optionally overlaying a local .env file.
func Load() (Config, error) {
	_ = loadDotEnv(".env")

	return Config{
		Port:            getEnv("PORT", "8080"),
		NodeAPIURL:      getEnv("NODE_API_URL", "http://localhost:5007"),
		NodeAPIUsername: getEnv("NODE_API_USERNAME", "admin@school-admin.com"),
		NodeAPIPassword: getEnv("NODE_API_PASSWORD", "3OU4zn3q6Zh9"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// loadDotEnv applies KEY=VALUE pairs from path into the process environment
// without overwriting values that are already set.
func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
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
			_ = os.Setenv(key, value)
		}
	}
	return scanner.Err()
}
