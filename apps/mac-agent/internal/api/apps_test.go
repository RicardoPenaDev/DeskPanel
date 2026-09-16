package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"deskpanel-agent/internal/appscan"
	"deskpanel-agent/internal/auth"
	"deskpanel-agent/internal/devices"
)

// writeFakeApp cria um .app mínimo mas com Info.plist real, suficiente
// para appscan.Scanner reconhecer via plutil.
func writeFakeApp(t *testing.T, dir, appName, bundleName string) {
	t.Helper()
	appPath := filepath.Join(dir, appName+".app")
	if err := os.MkdirAll(filepath.Join(appPath, "Contents"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	plist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key>
	<string>` + bundleName + `</string>
</dict>
</plist>
`
	if err := os.WriteFile(filepath.Join(appPath, "Contents", "Info.plist"), []byte(plist), 0o644); err != nil {
		t.Fatalf("WriteFile Info.plist: %v", err)
	}
}

func TestApps_RequiresToken(t *testing.T) {
	s, _, _ := newTestServer(t)
	s.Apps = appscan.New()
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, loopbackRequest("GET", "/api/v1/apps", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 sem token", rec.Code)
	}
}

func TestApps_ValidTokenListsLiveScan(t *testing.T) {
	dir := t.TempDir()
	writeFakeApp(t, dir, "Alpha", "Alpha App")
	writeFakeApp(t, dir, "Zeta", "Zeta App")

	s, store, _ := newTestServer(t)
	s.Apps = &appscan.Scanner{Dirs: []string{dir}}
	token := "token-de-teste"
	_ = store.Put(devices.Device{DeviceID: "dev-1", TokenHash: auth.HashToken(token)})

	r := loopbackRequest("GET", "/api/v1/apps", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, r)

	if rec.Code != 200 {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var list []appSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len(list) = %d, want 2", len(list))
	}
	if list[0].Name != "Alpha App" || list[1].Name != "Zeta App" {
		t.Fatalf("ordem/nome inesperados: %+v", list)
	}
	for _, a := range list {
		if a.HasIcon {
			t.Errorf("app %q não deveria reportar ícone (bundle de teste sem .icns)", a.Name)
		}
	}
}

func TestAppIcon_UnknownIDReturns404(t *testing.T) {
	s, store, _ := newTestServer(t)
	s.Apps = &appscan.Scanner{Dirs: []string{t.TempDir()}}
	token := "token-de-teste"
	_ = store.Put(devices.Device{DeviceID: "dev-1", TokenHash: auth.HashToken(token)})

	r := loopbackRequest("GET", "/api/v1/apps/app:0000000000000000/icon", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, r)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestAppIcon_RequiresToken(t *testing.T) {
	s, _, _ := newTestServer(t)
	s.Apps = appscan.New()
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, loopbackRequest("GET", "/api/v1/apps/app:abc/icon", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 sem token", rec.Code)
	}
}
