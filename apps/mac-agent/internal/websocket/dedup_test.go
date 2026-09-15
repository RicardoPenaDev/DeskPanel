package websocket

import (
	"testing"
	"time"
)

func TestDedupCache_GetMiss(t *testing.T) {
	d := newDedupCache(time.Minute)
	if _, ok := d.get("não-existe"); ok {
		t.Fatal("get() em cache vazio deveria retornar ok=false")
	}
}

func TestDedupCache_PutThenGet(t *testing.T) {
	d := newDedupCache(time.Minute)
	want := actionResultPayload{ActionID: "app.chrome", Status: "success"}
	d.put("req-1", want)

	got, ok := d.get("req-1")
	if !ok || got.ActionID != want.ActionID {
		t.Fatalf("get() = %+v, %v", got, ok)
	}
}

func TestDedupCache_ExpiresAfterTTL(t *testing.T) {
	d := newDedupCache(time.Minute)
	fakeNow := time.Now()
	d.now = func() time.Time { return fakeNow }

	d.put("req-1", actionResultPayload{ActionID: "app.chrome"})
	fakeNow = fakeNow.Add(2 * time.Minute)

	if _, ok := d.get("req-1"); ok {
		t.Fatal("get() deveria expirar depois do TTL")
	}
}
