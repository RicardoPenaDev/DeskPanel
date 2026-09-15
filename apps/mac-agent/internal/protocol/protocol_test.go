package protocol

import "testing"

func TestVersionIsPositive(t *testing.T) {
	if Version < 1 {
		t.Fatalf("protocol.Version = %d, want >= 1", Version)
	}
}

func TestMessageLimitsMatchSpec(t *testing.T) {
	if MaxWebSocketMessageBytes != 16*1024 {
		t.Errorf("MaxWebSocketMessageBytes = %d, want 16 KiB (docs/PROTOCOL.md)", MaxWebSocketMessageBytes)
	}
	if MaxHTTPRequestBytes != 64*1024 {
		t.Errorf("MaxHTTPRequestBytes = %d, want 64 KiB (docs/PROTOCOL.md)", MaxHTTPRequestBytes)
	}
}
