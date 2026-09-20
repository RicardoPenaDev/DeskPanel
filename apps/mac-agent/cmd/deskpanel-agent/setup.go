package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"deskpanel-agent/internal/adminsocket"
	qrcode "github.com/skip2/go-qrcode"
)

func cmdSetup(args []string) error {
	if _, err := adminsocket.Call(defaultAdminSocketPath(), "status", nil); err != nil {
		return fmt.Errorf("setup: agente não está rodando; inicie o LaunchAgent primeiro: %w", err)
	}
	page := setupHTML()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		ips := strings.Join(localIPv4s(), "</li><li>")
		_, _ = fmt.Fprintf(w, page, 38121, time.Now().Format("15:04"), ips)
	})
	mux.HandleFunc("POST /pair", func(w http.ResponseWriter, _ *http.Request) {
		result, err := adminsocket.Call(defaultAdminSocketPath(), "pair", nil)
		if err != nil {
			http.Error(w, "não foi possível gerar o código", http.StatusInternalServerError)
			return
		}
		var pairing pairResult
		if err := json.Unmarshal(result, &pairing); err != nil {
			http.Error(w, "resposta inválida do agente", http.StatusInternalServerError)
			return
		}
		ips := localIPv4s()
		host := ""
		if len(ips) > 0 {
			host = ips[0]
		}
		payload := "deskpanel://pair?host=" + url.QueryEscape(host) + "&port=38121&code=" + url.QueryEscape(pairing.Code) + "&protocolVersion=1"
		png, err := qrcode.Encode(payload, qrcode.Medium, 320)
		if err != nil {
			http.Error(w, "não foi possível gerar o QR Code", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"code":      pairing.Code,
			"qrDataUrl": "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
		})
	})

	server := &http.Server{Addr: "127.0.0.1:0", Handler: mux, ReadHeaderTimeout: 3 * time.Second}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("setup: %w", err)
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/", listener.Addr().(*net.TCPAddr).Port)
	if err := exec.Command("open", url).Start(); err != nil {
		_ = listener.Close()
		return fmt.Errorf("setup: não foi possível abrir o navegador: %w", err)
	}
	fmt.Printf("Assistente local aberto em %s\n", url)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { <-ctx.Done(); _ = server.Close() }()
	return server.Serve(listener)
}

func localIPv4s() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var result []string
	for _, iface := range interfaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil || ip.IsLoopback() || ip.To4() == nil || seen[ip.String()] {
				continue
			}
			seen[ip.String()] = true
			result = append(result, ip.String())
		}
	}
	return result
}

func setupHTML() string {
	return `<!doctype html><html lang="pt-BR"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>DeskPanel</title><style>
:root{color-scheme:dark;font-family:-apple-system,BlinkMacSystemFont,"SF Pro Display",system-ui,sans-serif;background:#0b0d10;color:#f5f5f7}body{margin:0;min-height:100vh;background:radial-gradient(circle at 20%% 10%%,#26385b,transparent 42%%),radial-gradient(circle at 90%% 20%%,#39264e,transparent 40%%),#0b0d10;display:grid;place-items:center}.card{width:min(680px,calc(100%% - 40px));padding:34px;border:1px solid #ffffff1c;border-radius:28px;background:#1c1f25b8;box-shadow:0 20px 70px #0008,inset 0 1px #ffffff1a;backdrop-filter:blur(24px)}h1{font-size:34px;margin:0 0 8px}p{color:#a9adb7;line-height:1.5}.grid{display:grid;grid-template-columns:1fr 1fr;gap:14px;margin:26px 0}.item{padding:18px;border-radius:18px;background:#ffffff0d;border:1px solid #ffffff12}.label{font-size:12px;color:#a9adb7;text-transform:uppercase;letter-spacing:.08em}.value{font-size:20px;margin-top:7px}button{border:0;border-radius:999px;padding:14px 20px;background:#4f8cff;color:#07101f;font-weight:700;font-size:16px;cursor:pointer}#code{font-size:40px;letter-spacing:.18em;color:#69d391;margin:20px 0;text-align:center}#result{text-align:center;margin-top:22px}#result img{display:block;margin:18px auto 0}.muted{font-size:13px}@media(max-width:560px){.grid{grid-template-columns:1fr}.card{padding:24px}}
</style><main class="card"><h1>DeskPanel</h1><p>Configure o painel Android para controlar este Mac pela rede local.</p><section class="grid"><div class="item"><div class="label">Porta</div><div class="value">%d</div></div><div class="item"><div class="label">Assistente aberto às</div><div class="value">%s</div></div></section><div class="item"><div class="label">Endereços do Mac</div><ul><li>%s</li></ul><p class="muted">Use um destes endereços no Android. O agente continua limitado à rede local.</p></div><p><button onclick="pair()">Gerar código e QR Code</button></p><div id="result"></div></main><script>async function pair(){let r=await fetch('/pair',{method:'POST'});let d=await r.json();document.querySelector('#result').innerHTML='<div class="label">Código válido por cinco minutos</div><div id="code">'+d.code+'</div><img alt="QR Code de pareamento" style="width:220px;height:220px;border-radius:16px;background:white;padding:10px" src="'+d.qrDataUrl+'"><p>Escaneie este QR Code no DeskPanel Android.</p>'}</script></html>`
}
