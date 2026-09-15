package auth

import "testing"

func TestHashToken_Deterministic(t *testing.T) {
	a := HashToken("token-de-teste")
	b := HashToken("token-de-teste")
	if a != b {
		t.Fatal("HashToken não é determinístico")
	}
}

func TestTokensMatch(t *testing.T) {
	hash := HashToken("segredo-correto")
	if !TokensMatch("segredo-correto", hash) {
		t.Error("TokensMatch deveria aceitar o token correto")
	}
	if TokensMatch("segredo-errado", hash) {
		t.Error("TokensMatch deveria rejeitar um token diferente")
	}
}
