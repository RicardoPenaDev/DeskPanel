package devices

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"deskpanel-agent/internal/auth"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "devices.json")
	s, err := LoadStore(path)
	if err != nil {
		t.Fatalf("LoadStore() erro inesperado: %v", err)
	}
	return s
}

func TestLoadStore_MissingFileIsEmpty(t *testing.T) {
	s := newTestStore(t)
	if len(s.List()) != 0 {
		t.Errorf("List() = %v, want vazio", s.List())
	}
}

func TestPutAndFind(t *testing.T) {
	s := newTestStore(t)
	d := Device{DeviceID: "dev-1", DeviceName: "Moto G60", TokenHash: auth.HashToken("segredo"), PairedAt: time.Now()}
	if err := s.Put(d); err != nil {
		t.Fatalf("Put() erro: %v", err)
	}

	got, ok := s.Find("dev-1")
	if !ok || got.DeviceName != "Moto G60" {
		t.Fatalf("Find() = %+v, %v", got, ok)
	}
}

func TestPut_RepairingReplacesEntry(t *testing.T) {
	s := newTestStore(t)
	_ = s.Put(Device{DeviceID: "dev-1", DeviceName: "Antigo", TokenHash: auth.HashToken("t1")})
	_ = s.Put(Device{DeviceID: "dev-1", DeviceName: "Novo", TokenHash: auth.HashToken("t2")})

	if len(s.List()) != 1 {
		t.Fatalf("List() tem %d entradas, want 1 (uma por dispositivo)", len(s.List()))
	}
	got, _ := s.Find("dev-1")
	if got.DeviceName != "Novo" {
		t.Errorf("DeviceName = %q, want %q", got.DeviceName, "Novo")
	}
}

func TestFindByToken(t *testing.T) {
	s := newTestStore(t)
	_ = s.Put(Device{DeviceID: "dev-1", TokenHash: auth.HashToken("token-certo")})

	if _, ok := s.FindByToken("token-certo"); !ok {
		t.Error("FindByToken() deveria aceitar o token correto")
	}
	if _, ok := s.FindByToken("token-errado"); ok {
		t.Error("FindByToken() deveria rejeitar um token diferente")
	}
}

func TestFindByToken_IgnoresRevokedDevices(t *testing.T) {
	s := newTestStore(t)
	_ = s.Put(Device{DeviceID: "dev-1", TokenHash: auth.HashToken("token-certo")})
	if err := s.Revoke("dev-1"); err != nil {
		t.Fatalf("Revoke() erro: %v", err)
	}

	if _, ok := s.FindByToken("token-certo"); ok {
		t.Error("FindByToken() não deveria aceitar o token de um dispositivo revogado")
	}
}

func TestRevoke_UnknownDevice(t *testing.T) {
	s := newTestStore(t)
	if err := s.Revoke("não-existe"); err == nil {
		t.Error("Revoke() deveria falhar para dispositivo desconhecido")
	}
}

func TestSave_PersistsAcrossReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devices.json")
	s1, _ := LoadStore(path)
	_ = s1.Put(Device{DeviceID: "dev-1", DeviceName: "Moto G60", TokenHash: auth.HashToken("t")})

	s2, err := LoadStore(path)
	if err != nil {
		t.Fatalf("LoadStore() (reload) erro: %v", err)
	}
	got, ok := s2.Find("dev-1")
	if !ok || got.DeviceName != "Moto G60" {
		t.Fatalf("dados não sobreviveram ao reload: %+v, %v", got, ok)
	}
}

func TestSave_FilePermissionsAreOwnerOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devices.json")
	s, _ := LoadStore(path)
	_ = s.Put(Device{DeviceID: "dev-1", TokenHash: auth.HashToken("t")})

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() erro: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("permissão de devices.json = %o, want 0600", perm)
	}
}

func TestSave_NeverWritesRawToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devices.json")
	s, _ := LoadStore(path)
	const rawToken = "token-super-secreto-nao-pode-vazar"
	_ = s.Put(Device{DeviceID: "dev-1", TokenHash: auth.HashToken(rawToken)})

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() erro: %v", err)
	}
	if contains(string(data), rawToken) {
		t.Error("devices.json contém o token em texto puro — deveria conter só o hash")
	}
}

func contains(haystack, needle string) bool {
	return len(needle) > 0 && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}

func TestSave_FixesPreExistingLoosePermissions(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatalf("Chmod() erro: %v", err)
	}
	path := filepath.Join(dir, "devices.json")

	s, _ := LoadStore(path)
	if err := s.Put(Device{DeviceID: "dev-1", TokenHash: auth.HashToken("t")}); err != nil {
		t.Fatalf("Put() erro: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat() erro: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("permissão do diretório = %o, want 0700 (deveria corrigir permissão pré-existente)", perm)
	}
}
