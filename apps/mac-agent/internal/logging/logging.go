// Package logging provides the agent's structured logger. Callers must
// never pass access tokens, full pairing codes or raw process output as
// fields (PROJECT.md §13).
package logging

import (
	"log/slog"
	"os"
)

// New returns a structured JSON logger writing to stdout.
func New() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler)
}
