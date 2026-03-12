package config

// Config contains process-level configuration for AgentFence components.
type Config struct {
	ListenAddress string
}

// Default returns a minimal default configuration suitable for local development.
func Default() Config {
	return Config{
		ListenAddress: ":8080",
	}
}
