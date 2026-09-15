package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("não foi possível escrever config de teste: %v", err)
	}
	return path
}

const validConfig = `{
  "schemaVersion": 1,
  "server": {"listenAddress": "0.0.0.0", "port": 38121, "pairingWindowSeconds": 300, "maxConnections": 5},
  "security": {"allowPublicNetworks": false, "requestsPerMinute": 120, "failedAuthLimit": 10},
  "actions": [{"id": "app.chrome", "label": "Chrome", "icon": "chrome", "kind": "open_app", "parameters": {"application": "Google Chrome"}}]
}`

func TestLoad_Valid(t *testing.T) {
	path := writeTempConfig(t, validConfig)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() erro inesperado: %v", err)
	}
	if cfg.Server.Port != 38121 {
		t.Errorf("Server.Port = %d, want 38121", cfg.Server.Port)
	}
	if len(cfg.Actions) != 1 {
		t.Errorf("len(Actions) = %d, want 1", len(cfg.Actions))
	}
}

func TestLoad_RejectsPublicNetworks(t *testing.T) {
	const cfg = `{
  "schemaVersion": 1,
  "server": {"listenAddress": "0.0.0.0", "port": 38121, "pairingWindowSeconds": 300, "maxConnections": 5},
  "security": {"allowPublicNetworks": true, "requestsPerMinute": 120, "failedAuthLimit": 10},
  "actions": []
}`
	path := writeTempConfig(t, cfg)
	if _, err := Load(path); err == nil {
		t.Fatal("Load() deveria rejeitar allowPublicNetworks=true")
	}
}

func TestLoad_RejectsInvalidPort(t *testing.T) {
	const cfg = `{
  "schemaVersion": 1,
  "server": {"listenAddress": "0.0.0.0", "port": 0, "pairingWindowSeconds": 300, "maxConnections": 5},
  "security": {"allowPublicNetworks": false, "requestsPerMinute": 120, "failedAuthLimit": 10},
  "actions": []
}`
	path := writeTempConfig(t, cfg)
	if _, err := Load(path); err == nil {
		t.Fatal("Load() deveria rejeitar porta 0")
	}
}

func TestLoad_RejectsDuplicateActionIDs(t *testing.T) {
	const cfg = `{
  "schemaVersion": 1,
  "server": {"listenAddress": "0.0.0.0", "port": 38121, "pairingWindowSeconds": 300, "maxConnections": 5},
  "security": {"allowPublicNetworks": false, "requestsPerMinute": 120, "failedAuthLimit": 10},
  "actions": [
    {"id": "app.chrome", "kind": "open_app"},
    {"id": "app.chrome", "kind": "open_url"}
  ]
}`
	path := writeTempConfig(t, cfg)
	if _, err := Load(path); err == nil {
		t.Fatal("Load() deveria rejeitar IDs de ação duplicados")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "não-existe.json")); err == nil {
		t.Fatal("Load() deveria falhar para arquivo inexistente")
	}
}
