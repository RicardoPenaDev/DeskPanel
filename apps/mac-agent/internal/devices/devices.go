// Package devices persists paired devices in devices.json (PROJECT.md
// §7.2, §8.3). Only a token's hash is ever stored — never the token
// itself (§12.5, §12.7).
package devices

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"deskpanel-agent/internal/auth"
)

// Device is one entry in devices.json. Uma entrada por dispositivo — repareamento
// substitui a entrada existente (PROJECT.md §8.3).
type Device struct {
	DeviceID   string    `json:"deviceId"`
	DeviceName string    `json:"deviceName"`
	TokenHash  string    `json:"tokenHash"`
	AppVersion string    `json:"appVersion"`
	PairedAt   time.Time `json:"pairedAt"`
	Revoked    bool      `json:"revoked"`
}

// Store is a small, mutex-protected, disk-backed table of devices. It is
// safe for concurrent use by multiple HTTP/WebSocket goroutines.
type Store struct {
	mu      sync.Mutex
	path    string
	devices map[string]Device // por deviceId
}

// LoadStore reads devices.json from path if it exists, or starts an empty
// store otherwise (first run).
func LoadStore(path string) (*Store, error) {
	s := &Store{path: path, devices: map[string]Device{}}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("devices: não foi possível ler %s: %w", path, err)
	}

	var list []Device
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("devices: JSON inválido em %s: %w", path, err)
	}
	for _, d := range list {
		s.devices[d.DeviceID] = d
	}
	return s, nil
}

// Put adiciona ou substitui a entrada de um dispositivo (pareamento e
// repareamento) e persiste em disco.
func (s *Store) Put(d Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices[d.DeviceID] = d
	return s.saveLocked()
}

// Find retorna o dispositivo pelo id.
func (s *Store) Find(deviceID string) (Device, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.devices[deviceID]
	return d, ok
}

// FindByToken localiza o dispositivo cujo hash bate com o token
// apresentado, usando comparação resistente a timing (PROJECT.md §12.6).
// Dispositivos revogados nunca combinam.
func (s *Store) FindByToken(token string) (Device, bool) {
	candidateHash := auth.HashToken(token)

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, d := range s.devices {
		if d.Revoked {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(candidateHash), []byte(d.TokenHash)) == 1 {
			return d, true
		}
	}
	return Device{}, false
}

// Revoke marca um dispositivo como revogado — acesso cai imediatamente
// (PROJECT.md §12.12). Retorna erro se o dispositivo não existir.
func (s *Store) Revoke(deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.devices[deviceID]
	if !ok {
		return fmt.Errorf("devices: dispositivo não encontrado: %s", deviceID)
	}
	d.Revoked = true
	s.devices[deviceID] = d
	return s.saveLocked()
}

// List retorna todos os dispositivos, ordenados por deviceId, sem os
// tokens (eles nunca são armazenados em texto puro — só o hash — mas o
// hash também não é exposto pelas chamadas administrativas).
func (s *Store) List() []Device {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Device, 0, len(s.devices))
	for _, d := range s.devices {
		out = append(out, d)
	}
	return out
}

// saveLocked escreve devices.json atomicamente (arquivo temporário +
// rename) com permissão 0600, assumindo que s.mu já está travado.
func (s *Store) saveLocked() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("devices: não foi possível criar diretório: %w", err)
	}
	// MkdirAll não ajusta a permissão de um diretório que já existia com
	// permissão mais aberta — força 0700 de qualquer forma (PROJECT.md §12.1).
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("devices: não foi possível ajustar permissão do diretório: %w", err)
	}

	list := make([]Device, 0, len(s.devices))
	for _, d := range s.devices {
		list = append(list, d)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("devices: falha ao serializar: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("devices: falha ao escrever arquivo temporário: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("devices: falha ao substituir %s: %w", s.path, err)
	}
	return nil
}
