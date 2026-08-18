package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config contains runtime settings for the Maestro application.
type Config struct {
	LLMAPIKey       string
	LLMModel        string
	LLMBaseURL      string
	Port            int
	MaestroPlansDir string
	CodebasePath    string
	TraceabilityURL string
}

// Load reads an optional .env file and returns configuration from the
// environment. Existing environment variables take precedence over .env.
func Load(envPath string) (Config, error) {
	if envPath == "" {
		envPath = ".env"
	}
	if err := loadDotEnv(envPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	port, err := strconv.Atoi(value("PORT", "8080"))
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a valid port number")
	}

	plansDir := value("MAESTRO_PLANS_DIR", "plans")
	codebasePath := value("CODEBASE_PATH", ".")
	if !filepath.IsAbs(codebasePath) {
		codebasePath, err = filepath.Abs(codebasePath)
		if err != nil {
			return Config{}, fmt.Errorf("resolve CODEBASE_PATH: %w", err)
		}
	}

	return Config{
		LLMAPIKey:       os.Getenv("LLM_API_KEY"),
		LLMModel:        value("LLM_MODEL", "gpt-4o-mini"),
		LLMBaseURL:      value("LLM_BASE_URL", "https://api.openai.com/v1"),
		Port:            port,
		MaestroPlansDir: plansDir,
		CodebasePath:    codebasePath,
		TraceabilityURL: os.Getenv("MAESTRO_TRACEABILITY_URL"),
	}, nil
}

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, raw, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) == "" {
			continue
		}
		key = strings.TrimSpace(key)
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		value := strings.TrimSpace(raw)
		value = strings.Trim(value, "\"'")
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set %s: %w", key, err)
		}
	}
	return scanner.Err()
}
