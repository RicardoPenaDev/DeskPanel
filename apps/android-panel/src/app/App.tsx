import { PROTOCOL_VERSION } from "../protocol/messages";

// Casca inicial do painel. As telas reais (primeiro acesso, painel
// principal, editor, configurações — PROJECT.md §10.2) chegam nas Fases
// 3 e 4; por enquanto isto só confirma que o app builda e roda.
export default function App() {
  return (
    <main style={{ display: "grid", placeItems: "center", height: "100%" }}>
      <div style={{ textAlign: "center" }}>
        <h1>DeskPanel</h1>
        <p style={{ color: "var(--dp-muted)" }}>
          Fase 0 — esqueleto do painel. Protocolo v{PROTOCOL_VERSION}.
        </p>
      </div>
    </main>
  );
}
