package pairing

import (
	"sync"
	"time"
)

// Estado de uma janela de pareamento (PROJECT.md §8.3): código de 6
// dígitos, expira em 5 minutos, uso único, no máximo 5 tentativas.
const DefaultMaxAttempts = 5

// Manager mantém no máximo uma janela de pareamento ativa por vez. Seguro
// para uso concorrente (chamado a partir de handlers HTTP e do socket
// administrativo).
type Manager struct {
	mu          sync.Mutex
	code        string
	expiresAt   time.Time
	attempts    int
	maxAttempts int
	active      bool
	now         func() time.Time // injetável nos testes
}

func NewManager() *Manager {
	return &Manager{maxAttempts: DefaultMaxAttempts, now: time.Now}
}

// Open gera um novo código e abre uma janela de pareamento válida por
// duration. Uma chamada a Open substitui qualquer janela anterior.
func (m *Manager) Open(duration time.Duration) (string, error) {
	code, err := GenerateCode()
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.code = code
	m.expiresAt = m.now().Add(duration)
	m.attempts = 0
	m.active = true
	return code, nil
}

// Close fecha a janela de pareamento ativa, se houver.
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.active = false
}

// Attempt valida um código submetido por um cliente. Um código correto
// fecha a janela (uso único). Um código incorreto conta como tentativa;
// depois de DefaultMaxAttempts tentativas erradas a janela também fecha.
func (m *Manager) Attempt(code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.active {
		return ErrClosed
	}
	if m.now().After(m.expiresAt) {
		m.active = false
		return ErrExpired
	}

	m.attempts++
	if code != m.code {
		if m.attempts >= m.maxAttempts {
			m.active = false
		}
		return ErrInvalidCode
	}

	m.active = false // uso único
	return nil
}
