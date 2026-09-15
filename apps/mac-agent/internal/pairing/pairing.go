// Package pairing generates the one-time codes used to authorize a new
// device (PROJECT.md §8.3). Session/expiry/attempt-limit tracking is Fase 2
// scope — this package only covers code generation for now.
package pairing

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// codeSpace is 10^6 — a 6-digit code (PROJECT.md §8.3).
var codeSpace = big.NewInt(1000000)

// GenerateCode returns a cryptographically random 6-digit pairing code,
// zero-padded.
func GenerateCode() (string, error) {
	n, err := rand.Int(rand.Reader, codeSpace)
	if err != nil {
		return "", fmt.Errorf("pairing: falha ao gerar código: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
