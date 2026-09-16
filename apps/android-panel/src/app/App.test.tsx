import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import App from "./App";
import { DEFAULT_APP_SETTINGS } from "../storage/appSettings";

const deviceControl = vi.hoisted(() => ({
  setKeepAwake: vi.fn(async () => {}),
  setImmersiveMode: vi.fn(async () => {}),
  setDimBrightness: vi.fn(async () => {}),
}));

vi.mock("../services/deviceControl", () => deviceControl);

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

  it("aplica keep-awake/imersivo/brilho assim que as preferências carregam, sem esperar a tela de Configurações", async () => {
    render(
      <App
        connectionOverrides={{
          loadConnectionConfig: vi.fn(async () => ({
            deviceId: "device-1",
            deviceName: "Moto G60",
            host: "",
            port: 38121,
          })),
          loadLayout: vi.fn(async () => ({
            schemaVersion: 1 as const,
            activeProfileId: "default",
            profiles: [],
          })),
          loadAppSettings: vi.fn(async () => DEFAULT_APP_SETTINGS),
          readAccessToken: vi.fn(async () => null),
        }}
      />,
    );

    await waitFor(() => expect(deviceControl.setKeepAwake).toHaveBeenCalledWith(true));
    expect(deviceControl.setImmersiveMode).toHaveBeenCalledWith(true);
    expect(deviceControl.setDimBrightness).toHaveBeenCalledWith(false);
  });
});
