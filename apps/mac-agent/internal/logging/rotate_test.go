package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRotatingWriterAppendsWithoutRotatingBelowLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.log")

	w, err := NewRotatingWriter(path, 1024, 3)
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	defer w.Close()

	if _, err := w.Write([]byte("linha 1\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := w.Write([]byte("linha 2\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "linha 1") || !strings.Contains(string(data), "linha 2") {
		t.Fatalf("conteúdo inesperado: %q", data)
	}
	if _, err := os.Stat(path + ".1"); !os.IsNotExist(err) {
		t.Fatalf("não deveria existir backup ainda: %v", err)
	}
}

func TestRotatingWriterRotatesPastMaxBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.log")

	// maxBytes pequeno o suficiente para forçar rotação numa única escrita.
	w, err := NewRotatingWriter(path, 10, 3)
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	defer w.Close()

	if _, err := w.Write([]byte("0123456789ab")); err != nil {
		t.Fatalf("Write 1: %v", err)
	}
	if _, err := w.Write([]byte("cdefgh")); err != nil {
		t.Fatalf("Write 2: %v", err)
	}

	if _, err := os.Stat(path + ".1"); err != nil {
		t.Fatalf("esperava backup .1 após ultrapassar maxBytes: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "cdefgh" {
		t.Fatalf("arquivo ativo deveria conter só a última escrita, veio: %q", data)
	}
}

func TestRotatingWriterKeepsAtMostMaxFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.log")

	w, err := NewRotatingWriter(path, 5, 3) // maxFiles=3 -> ativo + .1 + .2
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	defer w.Close()

	// Cada escrita de 6 bytes ultrapassa o limite de 5 e força rotação.
	for i := 0; i < 6; i++ {
		if _, err := w.Write([]byte("123456")); err != nil {
			t.Fatalf("Write %d: %v", i, err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) > 3 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("esperava no máximo 3 arquivos, achou %d: %v", len(entries), names)
	}
	if _, err := os.Stat(path + ".3"); !os.IsNotExist(err) {
		t.Fatalf("não deveria existir agent.log.3 (excede maxFiles=3): %v", err)
	}
}

func TestRotatingWriterCreatesFileIfMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "agent.log")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	w, err := NewRotatingWriter(path, 0, 0) // usa defaults
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	defer w.Close()

	if w.maxBytes != DefaultMaxBytes || w.maxFiles != DefaultMaxFiles {
		t.Fatalf("defaults não aplicados: maxBytes=%d maxFiles=%d", w.maxBytes, w.maxFiles)
	}
}

func TestRotatingWriterErrorsWhenDirectoryMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nao-existe", "agent.log")

	if _, err := NewRotatingWriter(path, 1024, 3); err == nil {
		t.Fatal("esperava erro ao abrir arquivo em diretório inexistente")
	}
}

func TestNewFileFallsBackToStdoutOnError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nao-existe", "agent.log")

	logger, closer, err := NewFile(path)
	if err == nil {
		t.Fatal("esperava erro reportado ao caller")
	}
	if logger == nil {
		t.Fatal("logger de fallback não deveria ser nil")
	}
	if closer != nil {
		t.Fatal("closer deveria ser nil no fallback para stdout")
	}
}
