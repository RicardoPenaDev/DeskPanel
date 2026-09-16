// Package corsorigins mantém a lista única de valores de Origin em que o
// DeskPanel Agent confia, compartilhada entre o handshake do WebSocket
// (internal/websocket) e o middleware CORS do servidor HTTP (internal/api).
// Ter uma lista só evita as duas ficarem fora de sincronia (PROJECT.md §12).
package corsorigins

// Allowed lista os valores de Origin que o WebView Android do Capacitor
// pode enviar. "localhost" sem esquema casa com qualquer protocolo
// (http/https/capacitor) desde que o host seja exatamente "localhost" —
// mesmo comportamento do OriginPatterns do pacote coder/websocket.
var Allowed = []string{"localhost", "https://localhost", "http://localhost", "capacitor://localhost"}
