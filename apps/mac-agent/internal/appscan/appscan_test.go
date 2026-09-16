package appscan

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeTestApp cria um bundle .app mínimo mas real (Info.plist válido)
// dentro de dir, opcionalmente copiando um .icns de origem para dentro
// dele — suficiente para plutil/sips reconhecerem de verdade.
func writeTestApp(t *testing.T, dir, appName, bundleName, iconFile, sourceIcns string) string {
	t.Helper()
	appPath := filepath.Join(dir, appName+".app")
	resources := filepath.Join(appPath, "Contents", "Resources")
	if err := os.MkdirAll(resources, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	plist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key>
	<string>` + bundleName + `</string>
	<key>CFBundleIconFile</key>
	<string>` + iconFile + `</string>
</dict>
</plist>
`
	if err := os.WriteFile(filepath.Join(appPath, "Contents", "Info.plist"), []byte(plist), 0o644); err != nil {
		t.Fatalf("WriteFile Info.plist: %v", err)
	}

	if sourceIcns != "" {
		data, err := os.ReadFile(sourceIcns)
		if err != nil {
			t.Fatalf("ReadFile %s: %v", sourceIcns, err)
		}
		name := iconFile
		if filepath.Ext(name) == "" {
			name += ".icns"
		}
		if err := os.WriteFile(filepath.Join(resources, name), data, 0o644); err != nil {
			t.Fatalf("WriteFile icon: %v", err)
		}
	}

	return appPath
}

// findSystemIcns procura um .icns real em /System/Applications para usar
// como fixture de conversão — evita empacotar um binário de teste.
func findSystemIcns(t *testing.T) string {
	t.Helper()
	matches, err := filepath.Glob("/System/Applications/*/Contents/Resources/*.icns")
	if err != nil || len(matches) == 0 {
		matches, _ = filepath.Glob("/System/Applications/Utilities/*/Contents/Resources/*.icns")
	}
	if len(matches) == 0 {
		t.Skip("nenhum .icns do sistema disponível neste ambiente")
	}
	return matches[0]
}

func TestScan_FindsAppsAndSortsByName(t *testing.T) {
	dir := t.TempDir()
	writeTestApp(t, dir, "Zeta", "Zeta App", "AppIcon", "")
	writeTestApp(t, dir, "Alpha", "Alpha App", "AppIcon", "")

	s := &Scanner{Dirs: []string{dir}}
	apps := s.Scan()

	if len(apps) != 2 {
		t.Fatalf("Scan() encontrou %d apps, esperava 2", len(apps))
	}
	if apps[0].Name != "Alpha App" || apps[1].Name != "Zeta App" {
		t.Fatalf("Scan() ordem inesperada: %q, %q", apps[0].Name, apps[1].Name)
	}
	if apps[0].ID == "" || apps[0].ID == apps[1].ID {
		t.Fatalf("Scan() IDs inválidos: %q, %q", apps[0].ID, apps[1].ID)
	}
}

func TestScan_IgnoresNonAppEntries(t *testing.T) {
	dir := t.TempDir()
	writeTestApp(t, dir, "RealApp", "Real App", "AppIcon", "")
	if err := os.WriteFile(filepath.Join(dir, "not-an-app.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "just-a-folder"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	s := &Scanner{Dirs: []string{dir}}
	apps := s.Scan()
	if len(apps) != 1 {
		t.Fatalf("Scan() encontrou %d apps, esperava 1 (ignorando não-apps)", len(apps))
	}
}

func TestScan_MissingDirIsSkipped(t *testing.T) {
	s := &Scanner{Dirs: []string{filepath.Join(t.TempDir(), "does-not-exist")}}
	apps := s.Scan()
	if len(apps) != 0 {
		t.Fatalf("Scan() esperava lista vazia para diretório inexistente, veio %d", len(apps))
	}
}

func TestResolve_FoundAndNotFound(t *testing.T) {
	dir := t.TempDir()
	writeTestApp(t, dir, "Alpha", "Alpha App", "AppIcon", "")

	s := &Scanner{Dirs: []string{dir}}
	apps := s.Scan()
	if len(apps) != 1 {
		t.Fatalf("setup: Scan() encontrou %d apps", len(apps))
	}

	if _, ok := s.Resolve(apps[0].ID); !ok {
		t.Fatal("Resolve() deveria encontrar o app recém-escaneado")
	}
	if _, ok := s.Resolve("app:0000000000000000"); ok {
		t.Fatal("Resolve() não deveria encontrar um ID inexistente")
	}
}

func TestHasIcon_FalseWhenIconFileMissing(t *testing.T) {
	dir := t.TempDir()
	writeTestApp(t, dir, "NoIcon", "No Icon App", "AppIcon", "")

	s := &Scanner{Dirs: []string{dir}}
	apps := s.Scan()
	if len(apps) != 1 {
		t.Fatalf("setup: Scan() encontrou %d apps", len(apps))
	}
	if apps[0].HasIcon() {
		t.Fatal("HasIcon() deveria ser false sem arquivo .icns no bundle")
	}
	if _, ok := s.Icon(apps[0].ID); ok {
		t.Fatal("Icon() deveria falhar para app sem ícone")
	}
}

func TestIcon_ConvertsRealIcnsToPNG(t *testing.T) {
	sourceIcns := findSystemIcns(t)

	dir := t.TempDir()
	writeTestApp(t, dir, "WithIcon", "With Icon App", "AppIcon", sourceIcns)

	s := &Scanner{Dirs: []string{dir}}
	apps := s.Scan()
	if len(apps) != 1 {
		t.Fatalf("setup: Scan() encontrou %d apps", len(apps))
	}
	if !apps[0].HasIcon() {
		t.Fatal("HasIcon() deveria ser true com .icns presente")
	}

	png, ok := s.Icon(apps[0].ID)
	if !ok {
		t.Fatal("Icon() deveria converter o .icns com sucesso")
	}
	if len(png) < 8 || string(png[1:4]) != "PNG" {
		t.Fatalf("Icon() não retornou um PNG válido (%d bytes)", len(png))
	}

	// Segunda chamada deve vir do cache e retornar o mesmo conteúdo.
	png2, ok := s.Icon(apps[0].ID)
	if !ok || string(png2) != string(png) {
		t.Fatal("Icon() em cache deveria retornar o mesmo PNG")
	}
}

func TestResolve_ReflectsAppRemovedFromDisk(t *testing.T) {
	dir := t.TempDir()
	appPath := writeTestApp(t, dir, "Ephemeral", "Ephemeral App", "AppIcon", "")

	s := &Scanner{Dirs: []string{dir}}
	apps := s.Scan()
	id := apps[0].ID

	if err := os.RemoveAll(appPath); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	if _, ok := s.Resolve(id); ok {
		t.Fatal("Resolve() não deveria encontrar um app removido do disco")
	}
}

// TestScan_CacheServesStaleResultWithinTTL prova a razão de existir do
// cache: dentro do ScanCacheTTL, um Scan() repetido não relê o disco —
// sem isso, N buscas de ícone quase simultâneas (o padrão real do
// Android ao abrir o painel) disparavam N variações inteiras de novo.
func TestScan_CacheServesStaleResultWithinTTL(t *testing.T) {
	dir := t.TempDir()
	appPath := writeTestApp(t, dir, "Cached", "Cached App", "AppIcon", "")

	s := &Scanner{Dirs: []string{dir}, ScanCacheTTL: time.Hour}
	first := s.Scan()
	if len(first) != 1 {
		t.Fatalf("setup: Scan() encontrou %d apps", len(first))
	}

	if err := os.RemoveAll(appPath); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	second := s.Scan()
	if len(second) != 1 {
		t.Fatal("Scan() deveria servir o cache (app removido não deveria ter sumido ainda)")
	}
}

func TestScan_CacheExpiresAfterTTL(t *testing.T) {
	dir := t.TempDir()
	appPath := writeTestApp(t, dir, "Expiring", "Expiring App", "AppIcon", "")

	s := &Scanner{Dirs: []string{dir}, ScanCacheTTL: 10 * time.Millisecond}
	if len(s.Scan()) != 1 {
		t.Fatal("setup: Scan() deveria encontrar 1 app")
	}

	if err := os.RemoveAll(appPath); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}
	time.Sleep(20 * time.Millisecond)

	if len(s.Scan()) != 0 {
		t.Fatal("Scan() deveria relevar o disco depois do TTL expirar")
	}
}
