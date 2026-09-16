// Package logging provides the agent's structured logger. Callers must
// never pass access tokens, full pairing codes or raw process output as
// fields (PROJECT.md §13). Log files rotate at ~5 MiB, keeping up to three
// files on disk (PROJECT.md §13).
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// New returns a structured JSON logger writing to stdout. Used in tests and
// when no file path is available.
func New() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler)
}

// NewFile returns a structured JSON logger writing to a rotating file at
// path (PROJECT.md §13: ~5 MiB per file, up to three files). If the file
// cannot be opened (missing directory, permission error), it falls back to
// stdout and returns the open error so the caller can report it — logging
// must never prevent the agent from starting or running.
func NewFile(path string) (*slog.Logger, io.Closer, error) {
	writer, err := NewRotatingWriter(path, DefaultMaxBytes, DefaultMaxFiles)
	if err != nil {
		return New(), nil, fmt.Errorf("logging: usando stdout, não foi possível abrir log em arquivo: %w", err)
	}
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler), writer, nil
}
