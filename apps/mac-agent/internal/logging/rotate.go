package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// RotatingWriter is an io.Writer that appends to a log file and rotates it
// once it grows past MaxBytes, keeping at most MaxFiles total on disk (the
// active file plus MaxFiles-1 backups — PROJECT.md §13: ~5 MiB per file, up
// to three files). Write failures (disk full, permission error, missing
// directory) are returned to the caller but never panic — the agent must
// keep running even if logging breaks (PROJECT.md §13 "falha de escrita do
// log não pode interromper o agente").
type RotatingWriter struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	maxFiles int
	file     *os.File
	size     int64
}

const (
	// DefaultMaxBytes is ~5 MiB per PROJECT.md §13.
	DefaultMaxBytes int64 = 5 * 1024 * 1024
	// DefaultMaxFiles keeps the active file plus two rotated backups.
	DefaultMaxFiles = 3
)

// NewRotatingWriter opens (creating if necessary) path for append and
// returns a writer that rotates it according to maxBytes/maxFiles. The
// parent directory must already exist.
func NewRotatingWriter(path string, maxBytes int64, maxFiles int) (*RotatingWriter, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	if maxFiles <= 0 {
		maxFiles = DefaultMaxFiles
	}
	w := &RotatingWriter{path: path, maxBytes: maxBytes, maxFiles: maxFiles}
	if err := w.openCurrent(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *RotatingWriter) openCurrent() error {
	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("logging: não foi possível abrir %s: %w", w.path, err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("logging: não foi possível checar %s: %w", w.path, err)
	}
	w.file = f
	w.size = info.Size()
	return nil
}

// Write implements io.Writer. It rotates the file first if the incoming
// write would push it past maxBytes.
func (w *RotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.openCurrent(); err != nil {
			return 0, err
		}
	}

	if w.size > 0 && w.size+int64(len(p)) > w.maxBytes {
		// Rotação é best-effort: se falhar (ex. permissão), seguimos
		// escrevendo no arquivo atual em vez de perder a mensagem de log.
		_ = w.rotateLocked()
	}

	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

// rotateLocked shifts path -> path.1 -> path.2 -> ... up to maxFiles-1
// backups, dropping the oldest, then reopens path fresh. Caller must hold
// w.mu.
func (w *RotatingWriter) rotateLocked() error {
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}

	numBackups := w.maxFiles - 1
	if numBackups < 1 {
		// maxFiles == 1: sem backups, só trunca reabrindo do zero.
		_ = os.Remove(w.path)
		return w.openCurrent()
	}

	// O backup mais antigo permitido sai de vez antes de deslocar os outros.
	_ = os.Remove(w.rotatedName(numBackups))
	for i := numBackups - 1; i >= 1; i-- {
		src := w.rotatedName(i)
		if _, err := os.Stat(src); err == nil {
			_ = os.Rename(src, w.rotatedName(i+1))
		}
	}
	if _, err := os.Stat(w.path); err == nil {
		if err := os.Rename(w.path, w.rotatedName(1)); err != nil {
			return err
		}
	}

	return w.openCurrent()
}

func (w *RotatingWriter) rotatedName(index int) string {
	dir := filepath.Dir(w.path)
	base := filepath.Base(w.path)
	return filepath.Join(dir, fmt.Sprintf("%s.%d", base, index))
}

// Close closes the underlying file.
func (w *RotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}
