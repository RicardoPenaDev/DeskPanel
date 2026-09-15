// Package actions defines the authoritative catalog of executable actions
// (PROJECT.md §7.4). The catalog is only ever configured on the Mac; the
// Android app receives IDs and metadata, never parameters it invented.
package actions

import "fmt"

// Kind enumerates the action kinds allowed in the MVP. Anything not listed
// here (shell, exec, free scripts, client-supplied args/paths/URLs) is out
// of scope by design — see PROJECT.md §7.4 and docs/SECURITY.md.
type Kind string

const (
	KindOpenApp        Kind = "open_app"
	KindOpenURL        Kind = "open_url"
	KindRunShortcut    Kind = "run_shortcut"
	KindKeystroke      Kind = "keystroke"
	KindVolumeDelta    Kind = "volume_delta"
	KindVolumeSet      Kind = "volume_set"
	KindMuteToggle     Kind = "mute_toggle"
	KindSpotifyControl Kind = "spotify_control"
	KindMusicControl   Kind = "music_control"
	KindScreenLock     Kind = "screen_lock"
	KindDisplaySleep   Kind = "display_sleep"
)

// knownKinds is used for validation; keep in sync with the const block above.
var knownKinds = map[Kind]bool{
	KindOpenApp: true, KindOpenURL: true, KindRunShortcut: true,
	KindKeystroke: true, KindVolumeDelta: true, KindVolumeSet: true,
	KindMuteToggle: true, KindSpotifyControl: true, KindMusicControl: true,
	KindScreenLock: true, KindDisplaySleep: true,
}

// Action is one entry from the authoritative catalog stored in config.json.
type Action struct {
	ID         string         `json:"id"`
	Label      string         `json:"label"`
	Icon       string         `json:"icon"`
	Kind       Kind           `json:"kind"`
	Parameters map[string]any `json:"parameters"`
}

// RequiresLongPress reports whether the Android UI must require a long
// press before sending this action (PROJECT.md §7.4, §10.2-C).
func (a Action) RequiresLongPress() bool {
	return a.Kind == KindScreenLock || a.Kind == KindDisplaySleep
}

// ValidateCatalog checks the invariants required before an agent will
// accept a catalog: unique IDs and known kinds (PROJECT.md §15.1).
func ValidateCatalog(catalog []Action) error {
	seen := make(map[string]bool, len(catalog))
	for _, a := range catalog {
		if a.ID == "" {
			return fmt.Errorf("actions: ação sem id")
		}
		if seen[a.ID] {
			return fmt.Errorf("actions: id duplicado: %q", a.ID)
		}
		seen[a.ID] = true

		if !knownKinds[a.Kind] {
			return fmt.Errorf("actions: kind desconhecido para %q: %q", a.ID, a.Kind)
		}
	}
	return nil
}
