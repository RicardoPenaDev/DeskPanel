// Package netguard checks that a client's IP belongs to a private or
// link-local range, as required by PROJECT.md §11/§12: o agente nunca deve
// aceitar conexões de um endereço público.
package netguard

import "net"

// IsPrivateOrLoopback reports whether ip is in a private (RFC 1918 / RFC
// 4193), link-local, or loopback range.
func IsPrivateOrLoopback(ip net.IP) bool {
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

// IsAllowedAddr faz o parse de um endereço "host:port" ou apenas "host" e
// aplica IsPrivateOrLoopback. Endereços que não são um IP válido (ex.:
// hostname) são rejeitados por padrão — o protocolo espera IPs de rede
// local, não DNS.
func IsAllowedAddr(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	ip := net.ParseIP(host)
	return IsPrivateOrLoopback(ip)
}
