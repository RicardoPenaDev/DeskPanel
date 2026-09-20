package executor

import (
	"context"
	"testing"

	"deskpanel-agent/internal/actions"
)

// recordedCall captures one invocation of a fake runCommandFunc so tests
// can assert exactly what MacOSExecutor would have run — without ever
// running it (PROJECT.md §15.1: testes automatizados não podem abrir
// aplicativos, mudar volume ou controlar mídia real).
type recordedCall struct {
	name string
	args []string
}

func fakeRunner(results map[string]commandResult, calls *[]recordedCall) runCommandFunc {
	return func(_ context.Context, name string, args ...string) commandResult {
		*calls = append(*calls, recordedCall{name: name, args: args})
		key := name
		if len(args) > 0 {
			key = name + " " + args[0]
		}
		if r, ok := results[key]; ok {
			return r
		}
		return commandResult{}
	}
}

func TestMacOSExecutor_OpenApp_Success(t *testing.T) {
	var calls []recordedCall
	m := &MacOSExecutor{run: fakeRunner(nil, &calls)}

	res := m.Execute(context.Background(), actions.Action{
		ID: "app.chrome", Kind: actions.KindOpenApp,
		Parameters: map[string]any{"application": "Google Chrome"},
	})

	if res.Status != "success" {
		t.Fatalf("Status = %q, want success (res=%+v)", res.Status, res)
	}
	if len(calls) != 1 || calls[0].name != "open" || calls[0].args[0] != "-a" || calls[0].args[1] != "Google Chrome" {
		t.Fatalf("chamada inesperada: %+v", calls)
	}
}

func TestMacOSExecutor_OpenApp_MissingParameter(t *testing.T) {
	var calls []recordedCall
	m := &MacOSExecutor{run: fakeRunner(nil, &calls)}

	res := m.Execute(context.Background(), actions.Action{ID: "app.x", Kind: actions.KindOpenApp})

	if res.Status != "error" || res.ErrorCode != "ACTION_FAILED" {
		t.Errorf("res = %+v, want status=error errorCode=ACTION_FAILED", res)
	}
	if len(calls) != 0 {
		t.Errorf("não deveria ter chamado o SO sem o parâmetro obrigatório, calls=%+v", calls)
	}
}

func TestMacOSExecutor_OpenApp_CommandFails(t *testing.T) {
	var calls []recordedCall
	run := func(_ context.Context, name string, args ...string) commandResult {
		calls = append(calls, recordedCall{name: name, args: args})
		return commandResult{err: errFake("aplicativo não encontrado")}
	}
	m := &MacOSExecutor{run: run}

	res := m.Execute(context.Background(), actions.Action{
		ID: "app.x", Kind: actions.KindOpenApp,
		Parameters: map[string]any{"application": "Não Existe"},
	})

	if res.Status != "error" || res.ErrorCode != "ACTION_FAILED" {
		t.Errorf("res = %+v, want status=error errorCode=ACTION_FAILED", res)
	}
}

func TestMacOSExecutor_OpenURL_AcceptsHTTPS(t *testing.T) {
	var calls []recordedCall
	m := &MacOSExecutor{run: fakeRunner(nil, &calls)}

	res := m.Execute(context.Background(), actions.Action{
		ID: "url.status", Kind: actions.KindOpenURL,
		Parameters: map[string]any{"url": "https://status.example.com"},
	})

	if res.Status != "success" {
		t.Fatalf("Status = %q, want success", res.Status)
	}
	if len(calls) != 1 || calls[0].args[0] != "https://status.example.com" {
		t.Fatalf("chamada inesperada: %+v", calls)
	}
}

func TestMacOSExecutor_OpenURL_RejectsNonHTTPScheme(t *testing.T) {
	cases := []string{"file:///etc/passwd", "javascript:alert(1)", "ftp://example.com", "not a url"}
	for _, raw := range cases {
		var calls []recordedCall
		m := &MacOSExecutor{run: fakeRunner(nil, &calls)}

		res := m.Execute(context.Background(), actions.Action{
			ID: "url.bad", Kind: actions.KindOpenURL,
			Parameters: map[string]any{"url": raw},
		})

		if res.Status != "error" || res.ErrorCode != "ACTION_NOT_ALLOWED" {
			t.Errorf("url=%q: res = %+v, want status=error errorCode=ACTION_NOT_ALLOWED", raw, res)
		}
		if len(calls) != 0 {
			t.Errorf("url=%q: não deveria ter chamado 'open', calls=%+v", raw, calls)
		}
	}
}

func TestMacOSExecutor_RunShortcut_Success(t *testing.T) {
	var calls []recordedCall
	m := &MacOSExecutor{run: fakeRunner(nil, &calls)}

	res := m.Execute(context.Background(), actions.Action{
		ID: "shortcut.work", Kind: actions.KindRunShortcut,
		Parameters: map[string]any{"shortcut": "Modo Trabalho"},
	})

	if res.Status != "success" {
		t.Fatalf("Status = %q, want success", res.Status)
	}
	if len(calls) != 1 || calls[0].name != "shortcuts" || calls[0].args[0] != "run" || calls[0].args[1] != "Modo Trabalho" {
		t.Fatalf("chamada inesperada: %+v", calls)
	}
}

func TestMacOSExecutor_VolumeDelta_ClampsAt100(t *testing.T) {
	var calls []recordedCall
	run := func(_ context.Context, name string, args ...string) commandResult {
		calls = append(calls, recordedCall{name: name, args: args})
		if len(args) > 0 && args[len(args)-1] == "output volume of (get volume settings)" {
			return commandResult{stdout: "95\n"}
		}
		return commandResult{}
	}
	m := &MacOSExecutor{run: run}

	res := m.Execute(context.Background(), actions.Action{
		ID: "volume.up", Kind: actions.KindVolumeDelta,
		Parameters: map[string]any{"delta": float64(20)},
	})

	if res.Status != "success" {
		t.Fatalf("Status = %q, want success (res=%+v)", res.Status, res)
	}
	if len(calls) != 2 {
		t.Fatalf("esperava 2 chamadas (ler + ajustar), calls=%+v", calls)
	}
	setCall := calls[1].args[len(calls[1].args)-1]
	if setCall != "set volume output volume 100" {
		t.Errorf("comando de ajuste = %q, want volume clampado em 100", setCall)
	}
}

func TestMacOSExecutor_VolumeDelta_ClampsAt0(t *testing.T) {
	var calls []recordedCall
	run := func(_ context.Context, name string, args ...string) commandResult {
		calls = append(calls, recordedCall{name: name, args: args})
		if len(args) > 0 && args[len(args)-1] == "output volume of (get volume settings)" {
			return commandResult{stdout: "5"}
		}
		return commandResult{}
	}
	m := &MacOSExecutor{run: run}

	res := m.Execute(context.Background(), actions.Action{
		ID: "volume.down", Kind: actions.KindVolumeDelta,
		Parameters: map[string]any{"delta": float64(-30)},
	})

	if res.Status != "success" {
		t.Fatalf("Status = %q, want success", res.Status)
	}
	setCall := calls[1].args[len(calls[1].args)-1]
	if setCall != "set volume output volume 0" {
		t.Errorf("comando de ajuste = %q, want volume clampado em 0", setCall)
	}
}

func TestMacOSExecutor_VolumeSet_ClampsToRange(t *testing.T) {
	var calls []recordedCall
	m := &MacOSExecutor{run: fakeRunner(nil, &calls)}

	res := m.Execute(context.Background(), actions.Action{
		ID: "volume.max", Kind: actions.KindVolumeSet,
		Parameters: map[string]any{"value": float64(500)},
	})

	if res.Status != "success" {
		t.Fatalf("Status = %q, want success", res.Status)
	}
	got := calls[0].args[len(calls[0].args)-1]
	if got != "set volume output volume 100" {
		t.Errorf("comando = %q, want volume clampado em 100", got)
	}
}

func TestMacOSExecutor_MuteToggle_NoParametersNeeded(t *testing.T) {
	var calls []recordedCall
	m := &MacOSExecutor{run: fakeRunner(nil, &calls)}

	res := m.Execute(context.Background(), actions.Action{ID: "mute", Kind: actions.KindMuteToggle})

	if res.Status != "success" {
		t.Fatalf("Status = %q, want success", res.Status)
	}
	if len(calls) != 1 {
		t.Fatalf("esperava 1 chamada, calls=%+v", calls)
	}
}

func TestMacOSExecutor_SpotifyControl_BuildsFixedScript(t *testing.T) {
	var calls []recordedCall
	m := &MacOSExecutor{run: fakeRunner(nil, &calls)}

	res := m.Execute(context.Background(), actions.Action{
		ID: "media.playpause", Kind: actions.KindSpotifyControl,
		Parameters: map[string]any{"operation": "play_pause"},
	})

	if res.Status != "success" {
		t.Fatalf("Status = %q, want success", res.Status)
	}
	script := calls[0].args[len(calls[0].args)-1]
	if script != `tell application "Spotify" to playpause` {
		t.Errorf("script = %q, formato inesperado", script)
	}
}

func TestMacOSExecutor_MusicControl_RejectsUnknownOperation(t *testing.T) {
	var calls []recordedCall
	m := &MacOSExecutor{run: fakeRunner(nil, &calls)}

	res := m.Execute(context.Background(), actions.Action{
		ID: "media.bad", Kind: actions.KindMusicControl,
		Parameters: map[string]any{"operation": "shuffle-all; rm -rf /"},
	})

	if res.Status != "error" || res.ErrorCode != "ACTION_NOT_ALLOWED" {
		t.Errorf("res = %+v, want status=error errorCode=ACTION_NOT_ALLOWED", res)
	}
	if len(calls) != 0 {
		t.Errorf("não deveria ter chamado osascript para uma operação desconhecida, calls=%+v", calls)
	}
}

func TestMacOSExecutor_DisplaySleep_UsesFixedCommand(t *testing.T) {
	var calls []recordedCall
	m := &MacOSExecutor{run: fakeRunner(nil, &calls)}

	res := m.Execute(context.Background(), actions.Action{ID: "system.display_sleep", Kind: actions.KindDisplaySleep})
	if res.Status != "success" {
		t.Fatalf("res = %+v, want success", res)
	}
	if len(calls) != 1 || calls[0].name != "pmset" || len(calls[0].args) != 1 || calls[0].args[0] != "displaysleepnow" {
		t.Fatalf("chamada inesperada: %+v", calls)
	}
}

func TestMacOSExecutor_ScreenLock_UsesFixedAppleScript(t *testing.T) {
	var calls []recordedCall
	m := &MacOSExecutor{run: fakeRunner(nil, &calls)}

	res := m.Execute(context.Background(), actions.Action{ID: "system.lock", Kind: actions.KindScreenLock})
	if res.Status != "success" {
		t.Fatalf("res = %+v, want success", res)
	}
	if len(calls) != 1 || calls[0].name != "osascript" || calls[0].args[0] != "-e" || calls[0].args[1] != `tell application "System Events" to keystroke "q" using {control down, command down}` {
		t.Fatalf("script/chamada inesperada: %+v", calls)
	}
}

func TestMacOSExecutor_Keystroke_UsesValidatedAppleScript(t *testing.T) {
	var calls []recordedCall
	m := &MacOSExecutor{run: fakeRunner(nil, &calls)}

	res := m.Execute(context.Background(), actions.Action{
		ID: "app.screenshot", Kind: actions.KindKeystroke,
		Parameters: map[string]any{
			"key":       "4",
			"modifiers": []any{"cmd", "shift"},
		},
	})

	if res.Status != "success" {
		t.Fatalf("res = %+v, want success", res)
	}
	if len(calls) != 1 || calls[0].name != "osascript" || calls[0].args[0] != "-e" {
		t.Fatalf("chamada inesperada: %+v", calls)
	}
	want := `tell application "System Events" to keystroke "4" using {command down, shift down}`
	if calls[0].args[1] != want {
		t.Errorf("script = %q, want %q", calls[0].args[1], want)
	}
}

func TestMacOSExecutor_Keystroke_RejectsUntrustedKeyAndModifier(t *testing.T) {
	cases := []map[string]any{
		{"key": "4; do shell script \"bad\"", "modifiers": []any{"cmd"}},
		{"key": "4", "modifiers": []any{"cmd", "command"}},
		{"key": "4", "modifiers": []any{"fn"}},
	}
	for _, parameters := range cases {
		var calls []recordedCall
		m := &MacOSExecutor{run: fakeRunner(nil, &calls)}
		res := m.Execute(context.Background(), actions.Action{
			ID: "shortcut.invalid", Kind: actions.KindKeystroke, Parameters: parameters,
		})
		if res.Status != "error" || res.ErrorCode != "ACTION_NOT_ALLOWED" {
			t.Errorf("parameters=%v: res=%+v, want ACTION_NOT_ALLOWED", parameters, res)
		}
		if len(calls) != 0 {
			t.Errorf("parameters=%v: não deveria chamar o SO, calls=%+v", parameters, calls)
		}
	}
}

// errFake é um erro mínimo só para os testes acima — sem depender de
// errors.New em cada caso.
type errFake string

func (e errFake) Error() string { return string(e) }
