package pairing

import (
	"errors"
	"testing"
	"time"
)

func TestManager_OpenThenAttempt_Success(t *testing.T) {
	m := NewManager()
	code, err := m.Open(5 * time.Minute)
	if err != nil {
		t.Fatalf("Open() erro: %v", err)
	}
	if err := m.Attempt(code); err != nil {
		t.Fatalf("Attempt(código correto) erro: %v", err)
	}
}

func TestManager_Attempt_WithoutOpenWindow(t *testing.T) {
	m := NewManager()
	if err := m.Attempt("123456"); !errors.Is(err, ErrClosed) {
		t.Fatalf("Attempt() = %v, want ErrClosed", err)
	}
}

func TestManager_Attempt_SingleUse(t *testing.T) {
	m := NewManager()
	code, _ := m.Open(5 * time.Minute)
	_ = m.Attempt(code)

	if err := m.Attempt(code); !errors.Is(err, ErrClosed) {
		t.Fatalf("segunda tentativa com o mesmo código = %v, want ErrClosed (uso único)", err)
	}
}

func TestManager_Attempt_WrongCode(t *testing.T) {
	m := NewManager()
	_, _ = m.Open(5 * time.Minute)
	if err := m.Attempt("000000"); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("Attempt(código errado) = %v, want ErrInvalidCode", err)
	}
}

func TestManager_Attempt_LocksAfterMaxAttempts(t *testing.T) {
	m := NewManager()
	code, _ := m.Open(5 * time.Minute)

	for i := 0; i < DefaultMaxAttempts; i++ {
		if err := m.Attempt("000000"); !errors.Is(err, ErrInvalidCode) {
			t.Fatalf("tentativa %d = %v, want ErrInvalidCode", i, err)
		}
	}

	// depois de esgotar as tentativas, nem o código certo deve mais funcionar.
	if err := m.Attempt(code); !errors.Is(err, ErrClosed) {
		t.Fatalf("após esgotar tentativas, Attempt(código certo) = %v, want ErrClosed", err)
	}
}

func TestManager_Attempt_Expired(t *testing.T) {
	m := NewManager()
	fakeNow := time.Now()
	m.now = func() time.Time { return fakeNow }

	code, _ := m.Open(5 * time.Minute)
	fakeNow = fakeNow.Add(6 * time.Minute)

	if err := m.Attempt(code); !errors.Is(err, ErrExpired) {
		t.Fatalf("Attempt() após expirar = %v, want ErrExpired", err)
	}
}

func TestManager_Open_ReplacesPreviousWindow(t *testing.T) {
	m := NewManager()
	oldCode, _ := m.Open(5 * time.Minute)
	newCode, _ := m.Open(5 * time.Minute)

	if oldCode == newCode {
		t.Skip("colisão de código aleatório — improvável, ignorando")
	}
	if err := m.Attempt(oldCode); !errors.Is(err, ErrInvalidCode) {
		t.Errorf("código da janela anterior deveria ser inválido, got %v", err)
	}
	if err := m.Attempt(newCode); err != nil {
		t.Errorf("código da janela atual deveria funcionar, got %v", err)
	}
}
