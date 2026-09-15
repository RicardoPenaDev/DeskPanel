// Package auth hashes and compares device access tokens. The Mac must only
// ever persist a token's hash, never the token itself (PROJECT.md §8.3,
// §12.6-12.7).
package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// HashToken returns the hex-encoded SHA-256 hash of an access token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// TokensMatch compares a candidate token against a stored hash using a
// timing-safe comparison (PROJECT.md §12.6).
func TokensMatch(candidateToken, storedHash string) bool {
	candidateHash := HashToken(candidateToken)
	return subtle.ConstantTimeCompare([]byte(candidateHash), []byte(storedHash)) == 1
}
