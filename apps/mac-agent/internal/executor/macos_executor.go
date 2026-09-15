package executor

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"deskpanel-agent/internal/actions"
)

// commandResult is what runCommandFunc returns — just enough for the
// dispatch logic below, never the whole os/exec.Cmd.
type commandResult struct {
	stdout string
	err    error
}

// runCommandFunc abstracts process execution so MacOSExecutor's dispatch
// logic can be unit-tested without ever touching the real OS: automated
// tests must not open apps, change volume or control media for real
// (PROJECT.md §15.1). MacOSExecutor uses realRunCommand by default; tests
// inject a fake.
type runCommandFunc func(ctx context.Context, name string, args ...string) commandResult

// realRunCommand is the only place in this package that actually spawns a
// process. It never goes through a shell (PROJECT.md §7.4/§12.2): name is
// resolved with exec.LookPath and args are passed as a slice, never
// concatenated into a string.
func realRunCommand(ctx context.Context, name string, args ...string) commandResult {
	path, err := exec.LookPath(name)
	if err != nil {
		return commandResult{err: fmt.Errorf("comando não encontrado: %s", name)}
	}
	cmd := exec.CommandContext(ctx, path, args...)
	out, err := cmd.Output()
	return commandResult{stdout: string(out), err: err}
}

// MacOSExecutor runs actions for real via exec.CommandContext. Only the
// action kinds listed in PROJECT.md §17 Fase 1 are implemented; the rest
// report ACTION_NOT_ALLOWED until their phase.
type MacOSExecutor struct {
	// run defaults to realRunCommand; tests override it.
	run runCommandFunc
}

func (m *MacOSExecutor) runner() runCommandFunc {
	if m.run != nil {
		return m.run
	}
	return realRunCommand
}

func (m *MacOSExecutor) Execute(ctx context.Context, action actions.Action) Result {
	start := time.Now()
	run := m.runner()

	var res Result
	switch action.Kind {
	case actions.KindOpenApp:
		res = execOpenApp(ctx, run, action)
	case actions.KindOpenURL:
		res = execOpenURL(ctx, run, action)
	case actions.KindRunShortcut:
		res = execRunShortcut(ctx, run, action)
	case actions.KindVolumeDelta:
		res = execVolumeDelta(ctx, run, action)
	case actions.KindVolumeSet:
		res = execVolumeSet(ctx, run, action)
	case actions.KindMuteToggle:
		res = execMuteToggle(ctx, run)
	case actions.KindSpotifyControl:
		res = execMediaControl(ctx, run, action, "Spotify")
	case actions.KindMusicControl:
		res = execMediaControl(ctx, run, action, "Music")
	default:
		res = Result{
			Status:    "error",
			ErrorCode: "ACTION_NOT_ALLOWED",
			Message:   fmt.Sprintf("kind %q ainda não implementado", action.Kind),
		}
	}

	if res.DurationMs == 0 {
		res.DurationMs = time.Since(start).Milliseconds()
	}
	return res
}

func stringParam(action actions.Action, key string) (string, bool) {
	v, ok := action.Parameters[key].(string)
	return v, ok && v != ""
}

// numberParam reads a numeric parameter. JSON numbers decode as float64 in
// map[string]any, so that's the only shape we accept here.
func numberParam(action actions.Action, key string) (int, bool) {
	v, ok := action.Parameters[key].(float64)
	if !ok {
		return 0, false
	}
	return int(v), true
}

func execOpenApp(ctx context.Context, run runCommandFunc, action actions.Action) Result {
	app, ok := stringParam(action, "application")
	if !ok {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: "parâmetro 'application' ausente"}
	}
	out := run(ctx, "open", "-a", app)
	if out.err != nil {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: fmt.Sprintf("não foi possível abrir %s", app)}
	}
	return Result{Status: "success"}
}

func execOpenURL(ctx context.Context, run runCommandFunc, action actions.Action) Result {
	raw, ok := stringParam(action, "url")
	if !ok {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: "parâmetro 'url' ausente"}
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		// PROJECT.md §7.4: open_url só aceita http/https.
		return Result{Status: "error", ErrorCode: "ACTION_NOT_ALLOWED", Message: "open_url só permite http/https"}
	}
	out := run(ctx, "open", raw)
	if out.err != nil {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: "não foi possível abrir a URL"}
	}
	return Result{Status: "success"}
}

func execRunShortcut(ctx context.Context, run runCommandFunc, action actions.Action) Result {
	name, ok := stringParam(action, "shortcut")
	if !ok {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: "parâmetro 'shortcut' ausente"}
	}
	out := run(ctx, "shortcuts", "run", name)
	if out.err != nil {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: fmt.Sprintf("não foi possível rodar o atalho %s", name)}
	}
	return Result{Status: "success"}
}

func clampVolume(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func setSystemVolume(ctx context.Context, run runCommandFunc, value int) Result {
	value = clampVolume(value)
	out := run(ctx, "osascript", "-e", fmt.Sprintf("set volume output volume %d", value))
	if out.err != nil {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: "não foi possível ajustar o volume"}
	}
	return Result{Status: "success"}
}

func execVolumeDelta(ctx context.Context, run runCommandFunc, action actions.Action) Result {
	delta, ok := numberParam(action, "delta")
	if !ok {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: "parâmetro 'delta' ausente"}
	}
	cur := run(ctx, "osascript", "-e", "output volume of (get volume settings)")
	if cur.err != nil {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: "não foi possível ler o volume atual"}
	}
	current, err := strconv.Atoi(strings.TrimSpace(cur.stdout))
	if err != nil {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: "resposta inesperada ao ler o volume"}
	}
	return setSystemVolume(ctx, run, current+delta)
}

func execVolumeSet(ctx context.Context, run runCommandFunc, action actions.Action) Result {
	value, ok := numberParam(action, "value")
	if !ok {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: "parâmetro 'value' ausente"}
	}
	return setSystemVolume(ctx, run, value)
}

func execMuteToggle(ctx context.Context, run runCommandFunc) Result {
	out := run(ctx, "osascript", "-e", "set volume output muted not (output muted of (get volume settings))")
	if out.err != nil {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: "não foi possível alternar o mudo"}
	}
	return Result{Status: "success"}
}

// mediaVerbs maps the three allowed client-facing operations to fixed
// AppleScript verbs — never built from client input (PROJECT.md §7.4).
var mediaVerbs = map[string]string{
	"play_pause": "playpause",
	"next":       "next track",
	"previous":   "previous track",
}

func execMediaControl(ctx context.Context, run runCommandFunc, action actions.Action, appName string) Result {
	op, ok := stringParam(action, "operation")
	if !ok {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: "parâmetro 'operation' ausente"}
	}
	verb, ok := mediaVerbs[op]
	if !ok {
		return Result{Status: "error", ErrorCode: "ACTION_NOT_ALLOWED", Message: fmt.Sprintf("operação %q não permitida", op)}
	}
	script := fmt.Sprintf("tell application %q to %s", appName, verb)
	out := run(ctx, "osascript", "-e", script)
	if out.err != nil {
		return Result{Status: "error", ErrorCode: "ACTION_FAILED", Message: fmt.Sprintf("não foi possível controlar %s", appName)}
	}
	return Result{Status: "success"}
}
