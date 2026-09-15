// Package executor runs actions against the real macOS or, in tests,
// against a fake that never touches the OS. Every action must have a
// bounded timeout and no child process may survive context cancellation
// (PROJECT.md §7.5).
package executor

import (
	"context"

	"deskpanel-agent/internal/actions"
)

// Result is what an Executor returns after attempting to run an action.
// Message is safe to show the client; internal details stay in the log
// (PROJECT.md §7.5, §13).
type Result struct {
	Status     string // "success" or "error"
	DurationMs int64
	ErrorCode  string
	Message    string
}

// Executor runs a single action and reports the outcome.
type Executor interface {
	Execute(ctx context.Context, action actions.Action) Result
}
