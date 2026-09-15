package actions

import "testing"

func TestValidateCatalog_OK(t *testing.T) {
	catalog := []Action{
		{ID: "app.chrome", Kind: KindOpenApp},
		{ID: "volume.up", Kind: KindVolumeDelta},
	}
	if err := ValidateCatalog(catalog); err != nil {
		t.Fatalf("ValidateCatalog() erro inesperado: %v", err)
	}
}

func TestValidateCatalog_DuplicateID(t *testing.T) {
	catalog := []Action{
		{ID: "app.chrome", Kind: KindOpenApp},
		{ID: "app.chrome", Kind: KindOpenURL},
	}
	if err := ValidateCatalog(catalog); err == nil {
		t.Fatal("ValidateCatalog() deveria rejeitar id duplicado")
	}
}

func TestValidateCatalog_UnknownKind(t *testing.T) {
	catalog := []Action{
		{ID: "shell.rm", Kind: Kind("shell")},
	}
	if err := ValidateCatalog(catalog); err == nil {
		t.Fatal("ValidateCatalog() deveria rejeitar kind desconhecido")
	}
}

func TestValidateCatalog_MissingID(t *testing.T) {
	catalog := []Action{{Kind: KindOpenApp}}
	if err := ValidateCatalog(catalog); err == nil {
		t.Fatal("ValidateCatalog() deveria rejeitar ação sem id")
	}
}

func TestRequiresLongPress(t *testing.T) {
	cases := []struct {
		kind Kind
		want bool
	}{
		{KindScreenLock, true},
		{KindDisplaySleep, true},
		{KindOpenApp, false},
	}
	for _, c := range cases {
		got := Action{Kind: c.kind}.RequiresLongPress()
		if got != c.want {
			t.Errorf("Action{Kind: %q}.RequiresLongPress() = %v, want %v", c.kind, got, c.want)
		}
	}
}
