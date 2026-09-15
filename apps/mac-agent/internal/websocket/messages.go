package websocket

import (
	"encoding/json"
	"time"

	"deskpanel-agent/internal/protocol"
)

type authHelloPayload struct {
	DeviceID    string `json:"deviceId"`
	AccessToken string `json:"accessToken"`
	AppVersion  string `json:"appVersion"`
}

type authAcceptedPayload struct {
	AgentVersion string `json:"agentVersion"`
	MacName      string `json:"macName"`
}

type actionExecutePayload struct {
	ActionID string `json:"actionId"`
}

type actionStartedPayload struct {
	ActionID string `json:"actionId"`
}

type actionResultPayload struct {
	ActionID   string  `json:"actionId"`
	Status     string  `json:"status"`
	DurationMs int64   `json:"durationMs"`
	ErrorCode  *string `json:"errorCode"`
	Message    *string `json:"message"`
}

type errorPayload struct {
	Code    protocol.ErrorCode `json:"code"`
	Message string             `json:"message"`
}

type stateSnapshotPayload struct {
	MacName      string `json:"macName"`
	AgentVersion string `json:"agentVersion"`
}

func marshalPayload(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		// só acontece com um payload mal formado pelo próprio agente — bug
		// de programação, não erro de runtime esperado.
		panic("websocket: falha ao serializar payload: " + err.Error())
	}
	return b
}

func newEnvelope(msgType, requestID string, payload any) protocol.Envelope {
	return protocol.Envelope{
		Type:            msgType,
		ProtocolVersion: protocol.Version,
		RequestID:       requestID,
		Timestamp:       time.Now().UTC(),
		Payload:         marshalPayload(payload),
	}
}
