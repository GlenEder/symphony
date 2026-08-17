package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	for _, key := range []string{"LLM_API_KEY", "PORT"} {
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("LLM_API_KEY=secret\nLLM_MODEL='test-model'\nPORT=9090\nLLM_BASE_URL=http://localhost:1234/v1\nMAESTRO_PLANS_DIR=/tmp/plans\nCODEBASE_PATH=/tmp/code\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLMAPIKey != "secret" || cfg.LLMModel != "test-model" || cfg.Port != 9090 || cfg.CodebasePath != "/tmp/code" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadDefaultsWhenEnvFileIsMissing(t *testing.T) {
	for _, key := range []string{"LLM_API_KEY", "LLM_MODEL", "LLM_BASE_URL", "PORT", "MAESTRO_PLANS_DIR", "CODEBASE_PATH"} {
		t.Setenv(key, "")
	}
	cfg, err := Load(filepath.Join(t.TempDir(), ".env"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 || cfg.LLMModel == "" || cfg.CodebasePath == "" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}
