package websocket

import (
	"sync"
	"time"
)

// atomicTime é um time.Time seguro para leitura/escrita concorrente — usado
// para o heartbeat trocar o "último pong visto" entre goroutines.
type atomicTime struct {
	mu sync.Mutex
	t  time.Time
}

func newAtomicTime(t time.Time) *atomicTime {
	return &atomicTime{t: t}
}

func (a *atomicTime) Get() time.Time {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.t
}

func (a *atomicTime) Set(t time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.t = t
}
