package websocket

import (
	"sync"
	"time"
)

// dedupCache lembra os requestId recentes para que um action.execute
// reenviado (ex.: depois de uma reconexão) não seja executado duas vezes
// (PROJECT.md §16).
type dedupCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	now     func() time.Time
	entries map[string]dedupEntry
}

type dedupEntry struct {
	result    actionResultPayload
	expiresAt time.Time
}

func newDedupCache(ttl time.Duration) *dedupCache {
	return &dedupCache{ttl: ttl, now: time.Now, entries: map[string]dedupEntry{}}
}

func (d *dedupCache) get(requestID string) (actionResultPayload, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	e, ok := d.entries[requestID]
	if !ok || d.now().After(e.expiresAt) {
		return actionResultPayload{}, false
	}
	return e.result, true
}

func (d *dedupCache) put(requestID string, result actionResultPayload) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.entries[requestID] = dedupEntry{result: result, expiresAt: d.now().Add(d.ttl)}
	d.sweepLocked()
}

func (d *dedupCache) sweepLocked() {
	now := d.now()
	for id, e := range d.entries {
		if now.After(e.expiresAt) {
			delete(d.entries, id)
		}
	}
}
