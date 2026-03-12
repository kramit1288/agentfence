package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultEnvironment = "development"
	defaultLogLevel    = "info"
	defaultLogFormat   = "json"
)

// Config contains process-level configuration for AgentFence components.
type Config struct {
	Environment string     `json:"environment"`
	Log         LogConfig  `json:"log"`
	HTTP        HTTPConfig `json:"http"`
}

// LogConfig controls process logging.
type LogConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// HTTPConfig controls the gateway HTTP server.
type HTTPConfig struct {
	Address           string        `json:"address"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	IdleTimeout       time.Duration `json:"idle_timeout"`
	ShutdownTimeout   time.Duration `json:"shutdown_timeout"`
}

// ValidationError aggregates configuration validation failures.
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid config: %s", strings.Join(e.Problems, "; "))
}

// Default returns a minimal default configuration suitable for local development.
func Default() Config {
	return Config{
		Environment: defaultEnvironment,
		Log: LogConfig{
			Level:  defaultLogLevel,
			Format: defaultLogFormat,
		},
		HTTP: HTTPConfig{
			Address:           ":8080",
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			ShutdownTimeout:   10 * time.Second,
		},
	}
}

// Validate checks whether the configuration is usable by the process.
func (c Config) Validate() error {
	var problems []string

	switch c.Environment {
	case "development", "staging", "production", "test":
	case "":
		problems = append(problems, "environment is required")
	default:
		problems = append(problems, "environment must be one of development, staging, production, test")
	}

	switch c.Log.Level {
	case "debug", "info", "warn", "error":
	case "":
		problems = append(problems, "log.level is required")
	default:
		problems = append(problems, "log.level must be one of debug, info, warn, error")
	}

	switch c.Log.Format {
	case "json", "text":
	case "":
		problems = append(problems, "log.format is required")
	default:
		problems = append(problems, "log.format must be one of json, text")
	}

	if c.HTTP.Address == "" {
		problems = append(problems, "http.address is required")
	}
	if c.HTTP.ReadHeaderTimeout <= 0 {
		problems = append(problems, "http.read_header_timeout must be greater than zero")
	}
	if c.HTTP.ReadTimeout <= 0 {
		problems = append(problems, "http.read_timeout must be greater than zero")
	}
	if c.HTTP.WriteTimeout <= 0 {
		problems = append(problems, "http.write_timeout must be greater than zero")
	}
	if c.HTTP.IdleTimeout <= 0 {
		problems = append(problems, "http.idle_timeout must be greater than zero")
	}
	if c.HTTP.ShutdownTimeout <= 0 {
		problems = append(problems, "http.shutdown_timeout must be greater than zero")
	}

	if len(problems) > 0 {
		return &ValidationError{Problems: problems}
	}

	return nil
}

// Load loads, merges, and validates configuration from defaults, file, and environment.
func Load(path string) (Config, error) {
	loader := loader{lookupEnv: os.LookupEnv}
	return loader.Load(path)
}

type loader struct {
	lookupEnv func(string) (string, bool)
}

func (l loader) Load(path string) (Config, error) {
	cfg := Default()

	if path != "" {
		fileConfig, err := loadFile(path)
		if err != nil {
			return Config{}, err
		}
		merge(&cfg, fileConfig)
	}

	if err := l.applyEnv(&cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadFile(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config file %q: %w", path, err)
	}

	return cfg, nil
}

func merge(dst *Config, src Config) {
	if src.Environment != "" {
		dst.Environment = src.Environment
	}
	if src.Log.Level != "" {
		dst.Log.Level = src.Log.Level
	}
	if src.Log.Format != "" {
		dst.Log.Format = src.Log.Format
	}
	if src.HTTP.Address != "" {
		dst.HTTP.Address = src.HTTP.Address
	}
	if src.HTTP.ReadHeaderTimeout != 0 {
		dst.HTTP.ReadHeaderTimeout = src.HTTP.ReadHeaderTimeout
	}
	if src.HTTP.ReadTimeout != 0 {
		dst.HTTP.ReadTimeout = src.HTTP.ReadTimeout
	}
	if src.HTTP.WriteTimeout != 0 {
		dst.HTTP.WriteTimeout = src.HTTP.WriteTimeout
	}
	if src.HTTP.IdleTimeout != 0 {
		dst.HTTP.IdleTimeout = src.HTTP.IdleTimeout
	}
	if src.HTTP.ShutdownTimeout != 0 {
		dst.HTTP.ShutdownTimeout = src.HTTP.ShutdownTimeout
	}
}

func (l loader) applyEnv(cfg *Config) error {
	if value, ok := l.lookupEnv("AGENTFENCE_ENVIRONMENT"); ok {
		cfg.Environment = value
	}
	if value, ok := l.lookupEnv("AGENTFENCE_LOG_LEVEL"); ok {
		cfg.Log.Level = value
	}
	if value, ok := l.lookupEnv("AGENTFENCE_LOG_FORMAT"); ok {
		cfg.Log.Format = value
	}
	if value, ok := l.lookupEnv("AGENTFENCE_HTTP_ADDRESS"); ok {
		cfg.HTTP.Address = value
	}
	if err := applyDurationEnv(l.lookupEnv, "AGENTFENCE_HTTP_READ_HEADER_TIMEOUT", &cfg.HTTP.ReadHeaderTimeout); err != nil {
		return err
	}
	if err := applyDurationEnv(l.lookupEnv, "AGENTFENCE_HTTP_READ_TIMEOUT", &cfg.HTTP.ReadTimeout); err != nil {
		return err
	}
	if err := applyDurationEnv(l.lookupEnv, "AGENTFENCE_HTTP_WRITE_TIMEOUT", &cfg.HTTP.WriteTimeout); err != nil {
		return err
	}
	if err := applyDurationEnv(l.lookupEnv, "AGENTFENCE_HTTP_IDLE_TIMEOUT", &cfg.HTTP.IdleTimeout); err != nil {
		return err
	}
	if err := applyDurationEnv(l.lookupEnv, "AGENTFENCE_HTTP_SHUTDOWN_TIMEOUT", &cfg.HTTP.ShutdownTimeout); err != nil {
		return err
	}

	return nil
}

func applyDurationEnv(lookupEnv func(string) (string, bool), key string, target *time.Duration) error {
	raw, ok := lookupEnv(key)
	if !ok {
		return nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("parse %s: %w", key, err)
	}

	*target = value
	return nil
}

// IsValidationError reports whether err contains a ValidationError.
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}
