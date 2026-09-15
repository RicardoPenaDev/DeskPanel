package executor

import (
	"context"
	"time"

	"deskpanel-agent/internal/actions"
)

// FakeExecutor never touches the real OS. It is what automated tests use so
// they cannot open apps, lock the Mac, change volume or control media for
// real (PROJECT.md §15.1).
type FakeExecutor struct {
	// Results, keyed by action ID, lets a test script a specific outcome.
	// A missing key falls back to a generic success result.
	Results map[string]Result
}

func (f *FakeExecutor) Execute(ctx context.Context, action actions.Action) Result {
	start := time.Now()
	if f.Results != nil {
		if r, ok := f.Results[action.ID]; ok {
			return r
		}
	}
	return Result{Status: "success", DurationMs: time.Since(start).Milliseconds()}
}
