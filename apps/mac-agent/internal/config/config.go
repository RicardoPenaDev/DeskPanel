// Package config loads and validates the agent's authoritative local
// configuration file (config.json — PROJECT.md §7.3). Loading must never
// silently accept an insecure or malformed catalog (PROJECT.md §19.2).
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"deskpanel-agent/internal/actions"
)

// Config mirrors configs/config.example.json.
type Config struct {
	SchemaVersion int              `json:"schemaVersion"`
	Server        Server           `json:"server"`
	Security      Security         `json:"security"`
	Actions       []actions.Action `json:"actions"`
}

type Server struct {
	ListenAddress        string `json:"listenAddress"`
	Port                 int    `json:"port"`
	PairingWindowSeconds int    `json:"pairingWindowSeconds"`
	MaxConnections       int    `json:"maxConnections"`
}

type Security struct {
	AllowPublicNetworks bool `json:"allowPublicNetworks"`
	RequestsPerMinute   int  `json:"requestsPerMinute"`
	FailedAuthLimit     int  `json:"failedAuthLimit"`
}

// Load reads and validates a config.json from disk.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: não foi possível ler %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: JSON inválido em %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate checks the invariants required before the agent will run with
// this configuration (PROJECT.md §12, §15.1).
func (c *Config) Validate() error {
	if c.SchemaVersion != 1 {
		return fmt.Errorf("config: schemaVersion não suportado: %d", c.SchemaVersion)
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("config: porta inválida: %d", c.Server.Port)
	}
	if c.Security.AllowPublicNetworks {
		return fmt.Errorf("config: allowPublicNetworks=true não é permitido no MVP (PROJECT.md §11-12)")
	}
	if err := actions.ValidateCatalog(c.Actions); err != nil {
		return err
	}
	return nil
}
