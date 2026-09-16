// Package appscan varre as pastas de aplicativos do próprio Mac
// (/Applications, /System/Applications) para oferecer ao Android uma
// lista sempre atualizada de apps instaláveis como atalho — sem depender
// de curadoria manual em config.json. Mantém a mesma regra de segurança
// das demais ações (PROJECT.md §7.4): o cliente nunca envia caminho,
// nome ou parâmetro livre — só escolhe um ID entre os que o próprio Mac
// já enumerou, e o Mac resolve esse ID de novo (Resolve) antes de
// executar qualquer coisa.
package appscan

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// defaultDirs são as pastas varridas — as mesmas onde o Finder/Launchpad
// procuram apps instalados.
var defaultDirs = []string{
	"/Applications",
	"/System/Applications",
	"/System/Applications/Utilities",
}

const (
	plutilPath   = "/usr/bin/plutil"
	sipsPath     = "/usr/bin/sips"
	scanTimeout  = 3 * time.Second
	iconTimeout  = 3 * time.Second
	iconMaxSizeP = 128 // px, lado maior
)

// App é um bundle .app encontrado em uma das pastas varridas.
type App struct {
	ID       string
	Name     string
	Path     string
	iconPath string
}

// HasIcon reporta se um ícone foi localizado dentro do bundle.
func (a App) HasIcon() bool {
	return a.iconPath != ""
}

type iconCacheEntry struct {
	modTime time.Time
	png     []byte
}

// Scanner enumera apps instalados e serve seus ícones como PNG. Toda
// leitura aqui é do próprio disco do Mac — nunca de um caminho recebido
// do Android.
type Scanner struct {
	Dirs []string

	mu    sync.Mutex
	icons map[string]iconCacheEntry
}

// New cria um Scanner com as pastas padrão.
func New() *Scanner {
	return &Scanner{}
}

func (s *Scanner) dirs() []string {
	if len(s.Dirs) > 0 {
		return s.Dirs
	}
	return defaultDirs
}

// Scan lista todo .app encontrado, ordenado por nome. Uma pasta ou
// bundle ilegível é ignorado silenciosamente — nunca derruba a lista
// inteira.
func (s *Scanner) Scan() []App {
	seen := make(map[string]bool)
	var apps []App

	for _, dir := range s.dirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".app") {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			app, ok := readBundle(path)
			if !ok || seen[app.ID] {
				continue
			}
			seen[app.ID] = true
			apps = append(apps, app)
		}
	}

	sort.Slice(apps, func(i, j int) bool {
		return strings.ToLower(apps[i].Name) < strings.ToLower(apps[j].Name)
	})
	return apps
}

// Resolve refaz a varredura e procura um app pelo ID. Usado logo antes
// de montar a ação a executar, para que um ID de app desinstalado nunca
// vire um "open_app" — sempre reflete o disco na hora.
func (s *Scanner) Resolve(id string) (App, bool) {
	for _, app := range s.Scan() {
		if app.ID == id {
			return app, true
		}
	}
	return App{}, false
}

// Icon devolve os bytes PNG do ícone de um app, convertendo do .icns
// original com `sips` (ferramenta nativa do macOS) e cacheando em
// memória por mtime do arquivo de origem.
func (s *Scanner) Icon(id string) ([]byte, bool) {
	app, ok := s.Resolve(id)
	if !ok || !app.HasIcon() {
		return nil, false
	}

	info, err := os.Stat(app.iconPath)
	if err != nil {
		return nil, false
	}

	s.mu.Lock()
	if cached, ok := s.icons[id]; ok && cached.modTime.Equal(info.ModTime()) {
		s.mu.Unlock()
		return cached.png, true
	}
	s.mu.Unlock()

	png, ok := convertIcon(app.iconPath)
	if !ok {
		return nil, false
	}

	s.mu.Lock()
	if s.icons == nil {
		s.icons = make(map[string]iconCacheEntry)
	}
	s.icons[id] = iconCacheEntry{modTime: info.ModTime(), png: png}
	s.mu.Unlock()
	return png, true
}

// appID deriva um ID estável e opaco do caminho do bundle — nunca expõe
// o caminho real do disco para o Android.
func appID(path string) string {
	sum := sha1.Sum([]byte(path))
	return "app:" + hex.EncodeToString(sum[:])[:16]
}

type bundlePlist struct {
	CFBundleName        string `json:"CFBundleName"`
	CFBundleDisplayName string `json:"CFBundleDisplayName"`
	CFBundleIconFile    string `json:"CFBundleIconFile"`
	CFBundleIconName    string `json:"CFBundleIconName"`
}

// readBundle lê Info.plist de um .app via `plutil` (nativo do macOS, lida
// com plist binário ou XML sem depender de biblioteca externa).
func readBundle(path string) (App, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()

	plistPath := filepath.Join(path, "Contents", "Info.plist")
	out, err := exec.CommandContext(ctx, plutilPath, "-convert", "json", "-o", "-", plistPath).Output()
	if err != nil {
		return App{}, false
	}

	var p bundlePlist
	if err := json.Unmarshal(out, &p); err != nil {
		return App{}, false
	}

	name := p.CFBundleDisplayName
	if name == "" {
		name = p.CFBundleName
	}
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(path), ".app")
	}

	return App{
		ID:       appID(path),
		Name:     name,
		Path:     path,
		iconPath: resolveIconPath(path, p),
	}, true
}

// resolveIconPath encontra o arquivo .icns real dentro do bundle.
// CFBundleIconFile/CFBundleIconName às vezes vêm sem a extensão.
func resolveIconPath(bundlePath string, p bundlePlist) string {
	for _, candidate := range []string{p.CFBundleIconFile, p.CFBundleIconName} {
		if candidate == "" {
			continue
		}
		if !strings.HasSuffix(candidate, ".icns") {
			candidate += ".icns"
		}
		full := filepath.Join(bundlePath, "Contents", "Resources", candidate)
		if _, err := os.Stat(full); err == nil {
			return full
		}
	}
	return ""
}

// convertIcon roda `sips` para converter um .icns em PNG, redimensionado
// para caber num ícone de tela — nunca via shell, args sempre fixos ou
// vindos do próprio disco do Mac.
func convertIcon(icnsPath string) ([]byte, bool) {
	dir, err := os.MkdirTemp("", "deskpanel-icon-*")
	if err != nil {
		return nil, false
	}
	defer os.RemoveAll(dir)
	out := filepath.Join(dir, "icon.png")

	ctx, cancel := context.WithTimeout(context.Background(), iconTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, sipsPath,
		"-s", "format", "png",
		"-Z", strconv.Itoa(iconMaxSizeP),
		icnsPath,
		"--out", out,
	)
	if err := cmd.Run(); err != nil {
		return nil, false
	}

	data, err := os.ReadFile(out)
	if err != nil {
		return nil, false
	}
	return data, true
}
