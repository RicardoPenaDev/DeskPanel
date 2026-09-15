package executor

import (
	"context"
	"testing"

	"deskpanel-agent/internal/actions"
)

func TestFakeExecutor_DefaultsToSuccess(t *testing.T) {
	f := &FakeExecutor{}
	res := f.Execute(context.Background(), actions.Action{ID: "app.chrome", Kind: actions.KindOpenApp})
	if res.Status != "success" {
		t.Errorf("Status = %q, want %q", res.Status, "success")
	}
}

func TestFakeExecutor_UsesScriptedResult(t *testing.T) {
	f := &FakeExecutor{Results: map[string]Result{
		"app.missing": {Status: "error", ErrorCode: "ACTION_FAILED", Message: "aplicativo não encontrado"},
	}}
	res := f.Execute(context.Background(), actions.Action{ID: "app.missing", Kind: actions.KindOpenApp})
	if res.Status != "error" || res.ErrorCode != "ACTION_FAILED" {
		t.Errorf("res = %+v, want status=error errorCode=ACTION_FAILED", res)
	}
}

func TestMacOSExecutor_NotImplementedYet(t *testing.T) {
	m := &MacOSExecutor{}
	res := m.Execute(context.Background(), actions.Action{ID: "app.chrome", Kind: actions.KindOpenApp})
	if res.Status != "error" {
		t.Errorf("Status = %q, want %q (Fase 1 ainda não implementada)", res.Status, "error")
	}
}
