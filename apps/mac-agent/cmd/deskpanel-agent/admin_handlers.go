package main

import (
	"encoding/json"
	"fmt"
	"time"

	"deskpanel-agent/internal/adminsocket"
	"deskpanel-agent/internal/config"
	"deskpanel-agent/internal/devices"
	"deskpanel-agent/internal/pairing"
	"deskpanel-agent/internal/websocket"
)

// deviceSummary é o que o comando "devices" expõe — nunca o hash do token
// (PROJECT.md §7.1 "devices: lista dispositivos pareados sem mostrar tokens").
type deviceSummary struct {
	DeviceID   string    `json:"deviceId"`
	DeviceName string    `json:"deviceName"`
	AppVersion string    `json:"appVersion"`
	PairedAt   time.Time `json:"pairedAt"`
	Revoked    bool      `json:"revoked"`
}

type pairResult struct {
	Code             string `json:"code"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
}

type statusResult struct {
	Port         int  `json:"port"`
	ActionsCount int  `json:"actionsCount"`
	Devices      int  `json:"devicesCount"`
	Connections  int  `json:"connections"`
	Running      bool `json:"running"`
}

// adminHandlers monta os handlers do socket administrativo (§7.1) usando o
// estado já criado por cmdServe.
func adminHandlers(cfg *config.Config, store *devices.Store, mgr *pairing.Manager, ws *websocket.Handler) map[string]adminsocket.HandlerFunc {
	return map[string]adminsocket.HandlerFunc{
		"pair": func(_ json.RawMessage) (any, error) {
			duration := time.Duration(cfg.Server.PairingWindowSeconds) * time.Second
			code, err := mgr.Open(duration)
			if err != nil {
				return nil, fmt.Errorf("falha ao gerar código de pareamento: %w", err)
			}
			return pairResult{Code: code, ExpiresInSeconds: cfg.Server.PairingWindowSeconds}, nil
		},

		"devices": func(_ json.RawMessage) (any, error) {
			list := store.List()
			out := make([]deviceSummary, 0, len(list))
			for _, d := range list {
				out = append(out, deviceSummary{
					DeviceID: d.DeviceID, DeviceName: d.DeviceName,
					AppVersion: d.AppVersion, PairedAt: d.PairedAt, Revoked: d.Revoked,
				})
			}
			return out, nil
		},

		"revoke": func(args json.RawMessage) (any, error) {
			var req struct {
				DeviceID string `json:"deviceId"`
			}
			if err := json.Unmarshal(args, &req); err != nil || req.DeviceID == "" {
				return nil, fmt.Errorf("deviceId é obrigatório")
			}
			if err := store.Revoke(req.DeviceID); err != nil {
				return nil, err
			}
			return map[string]string{"deviceId": req.DeviceID}, nil
		},

		"status": func(_ json.RawMessage) (any, error) {
			return statusResult{
				Port:         cfg.Server.Port,
				ActionsCount: len(cfg.Actions),
				Devices:      len(store.List()),
				Connections:  ws.ActiveConnections(),
				Running:      true,
			}, nil
		},
	}
}
