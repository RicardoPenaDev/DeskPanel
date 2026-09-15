// Package adminsocket implementa o socket Unix administrativo local que o
// deskpanel-agent serve expõe para os subcomandos pair/devices/revoke/status
// (PROJECT.md §7.1). Não aceita conexões de rede — só do mesmo usuário na
// mesma máquina, via arquivo de socket com permissão 0700 no diretório.
package adminsocket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Request é o que o cliente (CLI) manda para o socket.
type Request struct {
	Command string          `json:"command"`
	Args    json.RawMessage `json:"args,omitempty"`
}

// Response é o que o servidor (processo `serve`) devolve.
type Response struct {
	OK     bool            `json:"ok"`
	Error  string          `json:"error,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

// HandlerFunc processa um comando administrativo e retorna um valor
// serializável como resultado, ou um erro (nunca ambos).
type HandlerFunc func(args json.RawMessage) (any, error)

// Server escuta no socket Unix e despacha comandos para os handlers
// registrados.
type Server struct {
	SocketPath string
	Handlers   map[string]HandlerFunc
}

// Listen cria o socket em SocketPath (removendo um arquivo obsoleto de uma
// execução anterior que tenha encerrado sem limpar) com permissão 0700.
func (s *Server) Listen() (net.Listener, error) {
	dir := filepath.Dir(s.SocketPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("adminsocket: não foi possível criar diretório: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return nil, fmt.Errorf("adminsocket: não foi possível ajustar permissão do diretório: %w", err)
	}
	// Um socket obsoleto de uma execução anterior (crash sem cleanup)
	// impede o bind — removê-lo é seguro porque net.Listen já teria
	// falhado se outro processo estivesse escutando ativamente nele.
	_ = os.Remove(s.SocketPath)

	ln, err := net.Listen("unix", s.SocketPath)
	if err != nil {
		return nil, fmt.Errorf("adminsocket: não foi possível escutar em %s: %w", s.SocketPath, err)
	}
	if err := os.Chmod(s.SocketPath, 0o700); err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("adminsocket: não foi possível ajustar permissão: %w", err)
	}
	return ln, nil
}

// Serve aceita conexões até ln ser fechado. Bloqueia — chamar em uma goroutine.
func (s *Server) Serve(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	var req Request
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&req); err != nil {
		s.writeResponse(conn, Response{Error: "requisição inválida"})
		return
	}

	handler, ok := s.Handlers[req.Command]
	if !ok {
		s.writeResponse(conn, Response{Error: fmt.Sprintf("comando desconhecido: %s", req.Command)})
		return
	}

	result, err := handler(req.Args)
	if err != nil {
		s.writeResponse(conn, Response{Error: err.Error()})
		return
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		s.writeResponse(conn, Response{Error: "falha ao serializar resultado"})
		return
	}
	s.writeResponse(conn, Response{OK: true, Result: resultJSON})
}

func (s *Server) writeResponse(conn net.Conn, resp Response) {
	_ = json.NewEncoder(conn).Encode(resp)
}

// Call conecta em socketPath, manda um comando com args (serializado como
// JSON) e devolve o Result bruto. Erros de conexão viram uma mensagem
// explicando que o agente provavelmente não está rodando.
func Call(socketPath, command string, args any) (json.RawMessage, error) {
	conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("não foi possível conectar ao agente — ele está rodando? ('deskpanel-agent serve'): %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	var argsJSON json.RawMessage
	if args != nil {
		argsJSON, err = json.Marshal(args)
		if err != nil {
			return nil, fmt.Errorf("adminsocket: falha ao serializar argumentos: %w", err)
		}
	}

	if err := json.NewEncoder(conn).Encode(Request{Command: command, Args: argsJSON}); err != nil {
		return nil, fmt.Errorf("adminsocket: falha ao enviar comando: %w", err)
	}

	var resp Response
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil {
		return nil, fmt.Errorf("adminsocket: falha ao ler resposta: %w", err)
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	return resp.Result, nil
}
