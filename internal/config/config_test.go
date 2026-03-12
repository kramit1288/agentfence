package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfigIsValid(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatalf("Default() validation failed: %v", err)
	}
}

func TestLoadMergesFileAndEnv(t *testing.T) {
	t.Setenv("AGENTFENCE_LOG_LEVEL", "debug")
	t.Setenv("AGENTFENCE_HTTP_ADDRESS", ":9191")

	dir := t.TempDir()
	path := filepath.Join(dir, "agentfence.json")
	writeConfigFile(t, path, `{
		"environment": "staging",
		"log": {
			"format": "text"
		},
		"http": {
			"read_timeout": 30000000000,
			"write_timeout": 20000000000
		}
	}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Environment != "staging" {
		t.Fatalf("Environment = %q, want staging", cfg.Environment)
	}
	if cfg.Log.Level != "debug" {
		t.Fatalf("Log.Level = %q, want debug", cfg.Log.Level)
	}
	if cfg.Log.Format != "text" {
		t.Fatalf("Log.Format = %q, want text", cfg.Log.Format)
	}
	if cfg.HTTP.Address != ":9191" {
		t.Fatalf("HTTP.Address = %q, want :9191", cfg.HTTP.Address)
	}
	if cfg.HTTP.ReadTimeout != 30*time.Second {
		t.Fatalf("HTTP.ReadTimeout = %s, want 30s", cfg.HTTP.ReadTimeout)
	}
	if cfg.HTTP.WriteTimeout != 20*time.Second {
		t.Fatalf("HTTP.WriteTimeout = %s, want 20s", cfg.HTTP.WriteTimeout)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "read config file") {
		t.Fatalf("Load() error = %v, want read config file context", err)
	}
}

func TestLoadRejectsInvalidEnvDuration(t *testing.T) {
	t.Setenv("AGENTFENCE_HTTP_IDLE_TIMEOUT", "not-a-duration")

	_, err := Load("")
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "AGENTFENCE_HTTP_IDLE_TIMEOUT") {
		t.Fatalf("Load() error = %v, want env key in message", err)
	}
}

func TestLoadRejectsMissingRequiredFieldFromEnv(t *testing.T) {
	t.Setenv("AGENTFENCE_HTTP_ADDRESS", "")

	_, err := Load("")
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "http.address is required") {
		t.Fatalf("Load() error = %v, want missing address validation", err)
	}
}

func TestValidateRejectsMissingAndInvalidValues(t *testing.T) {
	cfg := Default()
	cfg.Environment = "invalid"
	cfg.Log.Level = ""
	cfg.Log.Format = "structured"
	cfg.HTTP.Address = ""
	cfg.HTTP.ReadHeaderTimeout = 0
	cfg.HTTP.ReadTimeout = -1 * time.Second

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
	if !IsValidationError(err) {
		t.Fatalf("Validate() error type = %T, want ValidationError", err)
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Validate() errors.As failed for %T", err)
	}
	if len(validationErr.Problems) < 4 {
		t.Fatalf("Validate() problems = %v, want multiple problems", validationErr.Problems)
	}
}

func writeConfigFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
}

