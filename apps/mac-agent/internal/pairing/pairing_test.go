package pairing

import "testing"

func TestGenerateCode_SixDigits(t *testing.T) {
	for i := 0; i < 20; i++ {
		code, err := GenerateCode()
		if err != nil {
			t.Fatalf("GenerateCode() erro: %v", err)
		}
		if len(code) != 6 {
			t.Fatalf("GenerateCode() = %q, want 6 dígitos", code)
		}
		for _, r := range code {
			if r < '0' || r > '9' {
				t.Fatalf("GenerateCode() = %q, contém caractere não numérico", code)
			}
		}
	}
}
