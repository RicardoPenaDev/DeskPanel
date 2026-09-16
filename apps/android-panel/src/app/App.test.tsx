import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import App from "./App";

describe("App", () => {
  it("mostra a tela de carregamento antes da configuração local terminar de carregar", () => {
    // Deps falsas (nunca resolvem) no lugar das reais baseadas em
    // @capacitor/preferences — este teste só verifica o estado síncrono
    // inicial, então não precisa (e não deve) tocar em storage de verdade.
    const neverResolves = () => new Promise<never>(() => {});
    render(
      <App
        connectionOverrides={{
          loadConnectionConfig: vi.fn(neverResolves),
          loadLayout: vi.fn(neverResolves),
          loadAppSettings: vi.fn(neverResolves),
        }}
      />,
    );
    expect(screen.getByText("DeskPanel")).toBeInTheDocument();
    expect(screen.getByText("Carregando…")).toBeInTheDocument();
  });
});
