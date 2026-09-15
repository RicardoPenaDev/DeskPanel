package netguard

import "testing"

func TestIsAllowedAddr(t *testing.T) {
	cases := []struct {
		addr string
		want bool
	}{
		{"192.168.1.10:54321", true},
		{"10.0.0.5:1234", true},
		{"172.16.4.4:1234", true},
		{"127.0.0.1:8080", true},
		{"169.254.1.1:8080", true}, // link-local
		{"::1", true},              // loopback IPv6
		{"8.8.8.8:443", false},     // público
		{"203.0.113.5:443", false}, // público (TEST-NET-3)
		{"não-é-ip:443", false},
	}
	for _, c := range cases {
		if got := IsAllowedAddr(c.addr); got != c.want {
			t.Errorf("IsAllowedAddr(%q) = %v, want %v", c.addr, got, c.want)
		}
	}
}
