package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfigPath_Shape(t *testing.T) {
	path := defaultConfigPath()
	if path == "" {
		t.Skip("HOME não definido neste ambiente")
	}
	want := filepath.Join("Library", "Application Support", "DeskPanel", "config.json")
	if !strings.HasSuffix(path, want) {
		t.Errorf("defaultConfigPath() = %q, want sufixo %q", path, want)
	}
}

func TestDefaultAppDir_IsParentOfConfigPath(t *testing.T) {
	dir := defaultAppDir()
	if dir == "" {
		t.Skip("HOME não definido neste ambiente")
	}
	if filepath.Join(dir, "config.json") != defaultConfigPath() {
		t.Errorf("defaultAppDir() = %q não é pai de defaultConfigPath() = %q", dir, defaultConfigPath())
	}
}

func TestDefaultDevicesAndSocketPaths_ShareAppDir(t *testing.T) {
	dir := defaultAppDir()
	if dir == "" {
		t.Skip("HOME não definido neste ambiente")
	}
	if filepath.Dir(defaultDevicesPath()) != dir {
		t.Errorf("defaultDevicesPath() não está em defaultAppDir()")
	}
	if filepath.Dir(defaultAdminSocketPath()) != dir {
		t.Errorf("defaultAdminSocketPath() não está em defaultAppDir()")
	}
}
