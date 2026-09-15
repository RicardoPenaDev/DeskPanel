package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"deskpanel-agent/internal/auth"
	"deskpanel-agent/internal/devices"
	"deskpanel-agent/internal/pairing"
	"deskpanel-agent/internal/protocol"
	"deskpanel-agent/internal/ratelimit"
)

type pairRequest struct {
	Code            string `json:"code"`
	DeviceID        string `json:"deviceId"`
	DeviceName      string `json:"deviceName"`
	AppVersion      string `json:"appVersion"`
	ProtocolVersion int    `json:"protocolVersion"`
}

type pairResponse struct {
	DeviceID        string `json:"deviceId"`
	AccessToken     string `json:"accessToken"`
	ProtocolVersion int    `json:"protocolVersion"`
}

// generateAccessToken gera um token com >= 32 bytes de entropia aleatória
// (PROJECT.md §8.3).
func generateAccessToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// PairHandler implementa POST /api/v1/pair (PROJECT.md §8.3). Não requer
// Bearer token — a prova de autorização aqui é o código temporário.
func PairHandler(mgr *pairing.Manager, store *devices.Store, failures *ratelimit.FailureTracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r)
		if failures.IsLocked(key) {
			writeError(w, http.StatusTooManyRequests, protocol.ErrRateLimited, "bloqueado temporariamente após tentativas repetidas")
			return
		}

		var req pairRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, protocol.ErrInvalidMessage, "JSON inválido")
			return
		}
		if req.ProtocolVersion != protocol.Version {
			writeError(w, http.StatusBadRequest, protocol.ErrProtocolUnsupported, "versão de protocolo não suportada")
			return
		}
		if req.DeviceID == "" || req.DeviceName == "" {
			writeError(w, http.StatusBadRequest, protocol.ErrInvalidMessage, "deviceId e deviceName são obrigatórios")
			return
		}

		if err := mgr.Attempt(req.Code); err != nil {
			failures.RecordFailure(key)
			switch {
			case errors.Is(err, pairing.ErrExpired):
				writeError(w, http.StatusForbidden, protocol.ErrPairingCodeExpired, "código de pareamento expirado")
			case errors.Is(err, pairing.ErrClosed):
				writeError(w, http.StatusForbidden, protocol.ErrPairingClosed, "pareamento não está aberto — rode 'deskpanel-agent pair' no Mac")
			default:
				writeError(w, http.StatusForbidden, protocol.ErrPairingCodeInvalid, "código de pareamento inválido")
			}
			return
		}
		failures.Reset(key)

		token, err := generateAccessToken()
		if err != nil {
			writeError(w, http.StatusInternalServerError, protocol.ErrInternalError, "falha ao gerar token")
			return
		}

		dev := devices.Device{
			DeviceID:   req.DeviceID,
			DeviceName: req.DeviceName,
			TokenHash:  auth.HashToken(token),
			AppVersion: req.AppVersion,
			PairedAt:   time.Now(),
		}
		if err := store.Put(dev); err != nil {
			writeError(w, http.StatusInternalServerError, protocol.ErrInternalError, "falha ao salvar dispositivo")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pairResponse{
			DeviceID:        req.DeviceID,
			AccessToken:     token,
			ProtocolVersion: protocol.Version,
		})
	}
}
