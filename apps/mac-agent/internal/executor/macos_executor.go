package executor

import (
	"context"

	"deskpanel-agent/internal/actions"
)

// MacOSExecutor will run actions for real via exec.CommandContext, respecting
// the restrictions in PROJECT.md §7.4 (no intermediary shell, no
// client-supplied arguments, paths or URLs). Dispatch by actions.Kind is
// Fase 1 scope — not implemented yet.
type MacOSExecutor struct{}

func (m *MacOSExecutor) Execute(ctx context.Context, action actions.Action) Result {
	return Result{
		Status:    "error",
		ErrorCode: "ACTION_NOT_ALLOWED",
		Message:   "MacOSExecutor ainda não implementado (Fase 1)",
	}
}
