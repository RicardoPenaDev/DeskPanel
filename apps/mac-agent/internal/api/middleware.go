package api

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"deskpanel-agent/internal/corsorigins"
	"deskpanel-agent/internal/devices"
	"deskpanel-agent/internal/netguard"
	"deskpanel-agent/internal/protocol"
	"deskpanel-agent/internal/ratelimit"
)

// originAllowed reporta se origin bate com algum valor de
// corsorigins.Allowed. Uma entrada sem esquema (ex. "localhost") casa
// com qualquer protocolo desde que o host seja exatamente igual — mesmo
// comportamento do OriginPatterns usado no handshake do WebSocket
// (internal/websocket/handler.go).
func originAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	for _, allowed := range corsorigins.Allowed {
		if strings.Contains(allowed, "://") {
			if origin == allowed {
				return true
			}
			continue
		}
		if u, err := url.Parse(origin); err == nil && u.Hostname() == allowed {
			return true
		}
	}
	return false
}

// cors adiciona os cabeçalhos necessários para o WebView Android do
// Capacitor (origem própria, ex. https://localhost) conseguir usar
// fetch()/XHR contra o agent, que está em outra origem
// (http://<ip-lan>:<porta>). Sem isso o WebView bloqueia a resposta antes
// de chegar ao código JS — o app enxerga isso como falha de rede genérica
// mesmo com a porta alcançável (mesma allowlist do WebSocket, PROJECT.md
// §12.13). Também responde o preflight OPTIONS, que o mux não trata.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if originAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "600")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// clientIP extrai só o endereço IP de r.RemoteAddr (sem a porta). Usado
// como chave de rate limit e para a checagem de rede privada.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// requirePrivateNetwork rejeita qualquer cliente fora de uma faixa
// privada/link-local/loopback (PROJECT.md §11, §12.16-17).
func requirePrivateNetwork(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !netguard.IsAllowedAddr(r.RemoteAddr) {
			writeError(w, http.StatusForbidden, protocol.ErrAuthInvalid, "acesso permitido só a partir da rede local")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimit aplica um limite de requisições por minuto por IP de origem
// (PROJECT.md §12.10).
func rateLimit(limiter *ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow(clientIP(r)) {
				writeError(w, http.StatusTooManyRequests, protocol.ErrRateLimited, "muitas requisições, tente novamente em instantes")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// maxBody limita o tamanho do corpo de qualquer requisição HTTP
// (PROJECT.md §12.5).
func maxBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, int64(protocol.MaxHTTPRequestBytes))
		next.ServeHTTP(w, r)
	})
}

// requireBearerToken exige "Authorization: Bearer <token>" e resolve o
// dispositivo correspondente. Um token ausente, inválido ou de um
// dispositivo revogado é sempre rejeitado — dispositivos revogados caem
// imediatamente porque devices.Store.FindByToken já os ignora
// (PROJECT.md §12.12).
func requireBearerToken(store *devices.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const prefix = "Bearer "
			header := r.Header.Get("Authorization")
			if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
				writeError(w, http.StatusUnauthorized, protocol.ErrAuthRequired, "token ausente")
				return
			}
			token := header[len(prefix):]

			dev, ok := store.FindByToken(token)
			if !ok {
				writeError(w, http.StatusUnauthorized, protocol.ErrAuthInvalid, "token inválido")
				return
			}

			next.ServeHTTP(w, r.WithContext(withDevice(r.Context(), dev)))
		})
	}
}
