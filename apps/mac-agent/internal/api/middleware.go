package api

import (
	"net"
	"net/http"

	"deskpanel-agent/internal/devices"
	"deskpanel-agent/internal/netguard"
	"deskpanel-agent/internal/protocol"
	"deskpanel-agent/internal/ratelimit"
)

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
