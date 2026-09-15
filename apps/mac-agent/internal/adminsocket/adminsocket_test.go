package adminsocket

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// shortSocketDir cria um diretório temporário curto — caminhos de socket
// Unix têm um limite de tamanho (~104-108 bytes no sun_path), e t.TempDir()
// aninhado no nome do teste facilmente estoura isso.
func shortSocketDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "dpsock")
	if err != nil {
		t.Fatalf("MkdirTemp() erro: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func startTestServer(t *testing.T, handlers map[string]HandlerFunc) string {
	t.Helper()
	path := filepath.Join(shortSocketDir(t), "agent.sock")
	s := &Server{SocketPath: path, Handlers: handlers}
	ln, err := s.Listen()
	if err != nil {
		t.Fatalf("Listen() erro: %v", err)
	}
	go s.Serve(ln)
	t.Cleanup(func() { _ = ln.Close() })
	return path
}

func TestCall_Success(t *testing.T) {
	path := startTestServer(t, map[string]HandlerFunc{
		"echo": func(args json.RawMessage) (any, error) {
			var payload struct {
				Msg string `json:"msg"`
			}
			_ = json.Unmarshal(args, &payload)
			return map[string]string{"echoed": payload.Msg}, nil
		},
	})

	result, err := Call(path, "echo", map[string]string{"msg": "oi"})
	if err != nil {
		t.Fatalf("Call() erro: %v", err)
	}
	var got map[string]string
	_ = json.Unmarshal(result, &got)
	if got["echoed"] != "oi" {
		t.Errorf("got %+v, want echoed=oi", got)
	}
}

func TestCall_HandlerError(t *testing.T) {
	path := startTestServer(t, map[string]HandlerFunc{
		"boom": func(args json.RawMessage) (any, error) {
			return nil, errors.New("algo deu errado")
		},
	})

	_, err := Call(path, "boom", nil)
	if err == nil || err.Error() != "algo deu errado" {
		t.Fatalf("Call() = %v, want erro 'algo deu errado'", err)
	}
}

func TestCall_UnknownCommand(t *testing.T) {
	path := startTestServer(t, map[string]HandlerFunc{})
	_, err := Call(path, "não-existe", nil)
	if err == nil {
		t.Fatal("Call() deveria falhar para comando desconhecido")
	}
}

func TestCall_NoServerRunning(t *testing.T) {
	path := filepath.Join(shortSocketDir(t), "não-existe.sock")
	_, err := Call(path, "status", nil)
	if err == nil {
		t.Fatal("Call() deveria falhar quando não há servidor escutando")
	}
}

func TestListen_SocketHasRestrictedPermissions(t *testing.T) {
	path := filepath.Join(shortSocketDir(t), "agent.sock")
	s := &Server{SocketPath: path, Handlers: map[string]HandlerFunc{}}
	ln, err := s.Listen()
	if err != nil {
		t.Fatalf("Listen() erro: %v", err)
	}
	defer ln.Close()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() erro: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("permissão do socket = %o, want 0700", perm)
	}
}

func TestListen_RemovesStaleSocketFile(t *testing.T) {
	path := filepath.Join(shortSocketDir(t), "agent.sock")
	if err := os.WriteFile(path, []byte("lixo de uma execução anterior"), 0o600); err != nil {
		t.Fatalf("WriteFile() erro: %v", err)
	}

	s := &Server{SocketPath: path, Handlers: map[string]HandlerFunc{}}
	ln, err := s.Listen()
	if err != nil {
		t.Fatalf("Listen() deveria remover o socket obsoleto e escutar normalmente, erro: %v", err)
	}
	defer ln.Close()
}
