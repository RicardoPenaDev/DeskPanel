// Package state models the snapshot the agent reports to a (re)connected
// client (state.snapshot / state.changed in docs/PROTOCOL.md). Real media
// and system status reporting is Fase 2+ scope.
package state

// Snapshot is the state sent to the Android app on connect/reconnect.
type Snapshot struct {
	MacName      string `json:"macName"`
	AgentVersion string `json:"agentVersion"`
}
