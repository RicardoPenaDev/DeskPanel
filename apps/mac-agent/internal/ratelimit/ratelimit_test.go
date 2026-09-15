package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter_AllowsUpToLimit(t *testing.T) {
	l := NewLimiter(3)
	for i := 0; i < 3; i++ {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("requisição %d deveria ser permitida", i)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Fatal("4ª requisição deveria ser bloqueada (limite=3)")
	}
}

func TestLimiter_SeparateKeysHaveSeparateBudgets(t *testing.T) {
	l := NewLimiter(1)
	if !l.Allow("a") || !l.Allow("b") {
		t.Fatal("chaves diferentes não deveriam compartilhar o orçamento")
	}
}

func TestLimiter_ResetsAfterWindow(t *testing.T) {
	l := NewLimiter(1)
	fakeNow := time.Now()
	l.now = func() time.Time { return fakeNow }

	if !l.Allow("k") {
		t.Fatal("primeira requisição deveria ser permitida")
	}
	if l.Allow("k") {
		t.Fatal("segunda requisição na mesma janela deveria ser bloqueada")
	}

	fakeNow = fakeNow.Add(time.Minute + time.Second)
	if !l.Allow("k") {
		t.Fatal("requisição na próxima janela deveria ser permitida")
	}
}

func TestLimiter_ZeroDisablesLimit(t *testing.T) {
	l := NewLimiter(0)
	for i := 0; i < 100; i++ {
		if !l.Allow("k") {
			t.Fatal("limite 0 deveria significar sem limite")
		}
	}
}

func TestFailureTracker_LocksAfterLimit(t *testing.T) {
	f := NewFailureTracker(3, time.Minute)
	for i := 0; i < 2; i++ {
		f.RecordFailure("dev-1")
		if f.IsLocked("dev-1") {
			t.Fatalf("não deveria estar bloqueado após %d falhas", i+1)
		}
	}
	f.RecordFailure("dev-1")
	if !f.IsLocked("dev-1") {
		t.Fatal("deveria estar bloqueado após atingir o limite")
	}
}

func TestFailureTracker_ResetClearsLock(t *testing.T) {
	f := NewFailureTracker(1, time.Minute)
	f.RecordFailure("dev-1")
	if !f.IsLocked("dev-1") {
		t.Fatal("deveria estar bloqueado")
	}
	f.Reset("dev-1")
	if f.IsLocked("dev-1") {
		t.Fatal("Reset() deveria limpar o bloqueio")
	}
}

func TestFailureTracker_UnlocksAfterDuration(t *testing.T) {
	f := NewFailureTracker(1, time.Minute)
	fakeNow := time.Now()
	f.now = func() time.Time { return fakeNow }

	f.RecordFailure("dev-1")
	if !f.IsLocked("dev-1") {
		t.Fatal("deveria estar bloqueado")
	}

	fakeNow = fakeNow.Add(2 * time.Minute)
	if f.IsLocked("dev-1") {
		t.Fatal("deveria ter desbloqueado depois do tempo de lockout")
	}
}
