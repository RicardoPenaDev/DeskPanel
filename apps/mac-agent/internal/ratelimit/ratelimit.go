// Package ratelimit implements the two throttles required by
// PROJECT.md §12.10-11: a requests-per-minute cap per key (IP ou
// dispositivo) e um bloqueio temporário depois de falhas de autenticação
// repetidas.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter is a fixed-window rate limiter keyed by an arbitrary string
// (normalmente um IP). Uma janela nova começa na primeira requisição depois
// da anterior expirar.
type Limiter struct {
	mu       sync.Mutex
	perMin   int
	window   time.Duration
	now      func() time.Time
	counters map[string]*window
}

type window struct {
	count   int
	resetAt time.Time
}

// NewLimiter cria um limitador que aceita até perMin requisições por
// minuto e por chave. perMin <= 0 desativa o limite (sempre permite).
func NewLimiter(perMin int) *Limiter {
	return &Limiter{
		perMin:   perMin,
		window:   time.Minute,
		now:      time.Now,
		counters: map[string]*window{},
	}
}

// Allow reports whether a request for key is within the limit, and conta
// a requisição atual.
func (l *Limiter) Allow(key string) bool {
	if l.perMin <= 0 {
		return true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	w, ok := l.counters[key]
	if !ok || now.After(w.resetAt) {
		w = &window{count: 0, resetAt: now.Add(l.window)}
		l.counters[key] = w
	}

	w.count++
	return w.count <= l.perMin
}

// FailureTracker bloqueia uma chave temporariamente depois de um número
// de falhas consecutivas de autenticação (PROJECT.md §12.11).
type FailureTracker struct {
	mu       sync.Mutex
	limit    int
	lockFor  time.Duration
	now      func() time.Time
	failures map[string]int
	lockedAt map[string]time.Time
}

func NewFailureTracker(limit int, lockFor time.Duration) *FailureTracker {
	return &FailureTracker{
		limit:    limit,
		lockFor:  lockFor,
		now:      time.Now,
		failures: map[string]int{},
		lockedAt: map[string]time.Time{},
	}
}

// RecordFailure conta mais uma falha para a chave; ao atingir o limite, a
// chave fica bloqueada por lockFor.
func (f *FailureTracker) RecordFailure(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failures[key]++
	if f.failures[key] >= f.limit {
		f.lockedAt[key] = f.now()
	}
}

// Reset limpa o contador de uma chave (chamar em toda autenticação bem-sucedida).
func (f *FailureTracker) Reset(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.failures, key)
	delete(f.lockedAt, key)
}

// IsLocked reports whether key está temporariamente bloqueada.
func (f *FailureTracker) IsLocked(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	lockedAt, ok := f.lockedAt[key]
	if !ok {
		return false
	}
	if f.now().Sub(lockedAt) >= f.lockFor {
		delete(f.lockedAt, key)
		delete(f.failures, key)
		return false
	}
	return true
}
