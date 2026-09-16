import { afterEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import SettingsScreen from "./SettingsScreen";
import { DEFAULT_APP_SETTINGS } from "../../storage/appSettings";

const deviceControl = vi.hoisted(() => ({
  setKeepAwake: vi.fn(async () => {}),
  setImmersiveMode: vi.fn(async () => {}),
  setDimBrightness: vi.fn(async () => {}),
}));

vi.mock("../../services/deviceControl", () => deviceControl);

const baseConfig = {
  deviceId: "device-1",
  deviceName: "Moto G60 do Ricardo",
  host: "192.168.1.50",
  port: 38121,
};

function renderScreen(overrides: Partial<Parameters<typeof SettingsScreen>[0]> = {}) {
  const props = {
    connectionConfig: baseConfig,
    appSettings: DEFAULT_APP_SETTINGS,
    status: "connected" as const,
    macName: "Mac do Ricardo",
    agentVersion: "0.1.0",
    connectionError: null,
    testConnection: vi.fn(async () => ({
      ok: true as const,
      data: { status: "ok", service: "deskpanel-agent", protocolVersion: 1 },
    })),
    updateConnectionConfig: vi.fn(async () => {}),
    updateAppSettings: vi.fn(async () => {}),
    reconnect: vi.fn(),
    forgetPairing: vi.fn(async () => {}),
    onClose: vi.fn(),
    ...overrides,
  };
  render(<SettingsScreen {...props} />);
  return props;
}

describe("SettingsScreen", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("mostra os campos de conexão preenchidos com a config atual", () => {
    renderScreen();
    expect(screen.getByDisplayValue("Moto G60 do Ricardo")).toBeInTheDocument();
    expect(screen.getByDisplayValue("192.168.1.50")).toBeInTheDocument();
    expect(screen.getByDisplayValue("38121")).toBeInTheDocument();
  });

  it("salva a configuração de conexão editada", async () => {
    const props = renderScreen();
    fireEvent.change(screen.getByDisplayValue("192.168.1.50"), {
      target: { value: "192.168.1.99" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Salvar" }));

    await waitFor(() =>
      expect(props.updateConnectionConfig).toHaveBeenCalledWith({
        ...baseConfig,
        host: "192.168.1.99",
      }),
    );
  });

  it("chama reconnect ao clicar em Reconectar", () => {
    const props = renderScreen();
    fireEvent.click(screen.getByRole("button", { name: "Reconectar" }));
    expect(props.reconnect).toHaveBeenCalledTimes(1);
  });

  it("pede confirmação antes de refazer o pareamento", async () => {
    const props = renderScreen();
    fireEvent.click(screen.getByRole("button", { name: "Refazer pareamento" }));
    expect(screen.getByText(/apaga o token salvo/)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Confirmar" }));
    await waitFor(() => expect(props.forgetPairing).toHaveBeenCalledTimes(1));
  });

  it("cancelar a confirmação não chama forgetPairing", () => {
    const props = renderScreen();
    fireEvent.click(screen.getByRole("button", { name: "Refazer pareamento" }));
    fireEvent.click(screen.getByRole("button", { name: "Cancelar" }));
    expect(props.forgetPairing).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Refazer pareamento" })).toBeInTheDocument();
  });

  it("liga vibração: atualiza appSettings sem tocar no plugin nativo", async () => {
    const props = renderScreen();
    fireEvent.click(screen.getByLabelText(/Vibração/));

    await waitFor(() =>
      expect(props.updateAppSettings).toHaveBeenCalledWith({
        ...DEFAULT_APP_SETTINGS,
        vibrationEnabled: false,
      }),
    );
    expect(deviceControl.setKeepAwake).not.toHaveBeenCalled();
  });

  it("desliga manter tela ligada: atualiza appSettings e chama o plugin nativo", async () => {
    renderScreen();
    fireEvent.click(screen.getByLabelText(/Manter tela ligada/));

    await waitFor(() => expect(deviceControl.setKeepAwake).toHaveBeenCalledWith(false));
  });

  it("liga modo imersivo: chama o plugin nativo correspondente", async () => {
    // immersiveModeEnabled já vem ligado por padrão (painel fixo na
    // parede) — parte de desligado aqui pra testar o caminho de "ligar".
    renderScreen({ appSettings: { ...DEFAULT_APP_SETTINGS, immersiveModeEnabled: false } });
    fireEvent.click(screen.getByLabelText(/Modo imersivo/));

    await waitFor(() => expect(deviceControl.setImmersiveMode).toHaveBeenCalledWith(true));
  });

  it("desliga modo imersivo: chama o plugin nativo correspondente", async () => {
    renderScreen();
    fireEvent.click(screen.getByLabelText(/Modo imersivo/));

    await waitFor(() => expect(deviceControl.setImmersiveMode).toHaveBeenCalledWith(false));
  });

  it("liga brilho reduzido: chama o plugin nativo correspondente", async () => {
    renderScreen();
    fireEvent.click(screen.getByLabelText(/Brilho reduzido/));

    await waitFor(() => expect(deviceControl.setDimBrightness).toHaveBeenCalledWith(true));
  });

  it("roda o diagnóstico de conexão e mostra o resultado", async () => {
    renderScreen();
    fireEvent.click(screen.getByRole("button", { name: "Testar conexão agora" }));

    await waitFor(() =>
      expect(screen.getByText(/deskpanel-agent, protocolo v1/)).toBeInTheDocument(),
    );
  });

  it("mostra erro do diagnóstico quando a checagem falha", async () => {
    renderScreen({
      testConnection: vi.fn(async () => ({
        ok: false as const,
        error: { code: "NETWORK_ERROR" as const, message: "Não foi possível alcançar o Mac" },
      })),
    });
    fireEvent.click(screen.getByRole("button", { name: "Testar conexão agora" }));

    await waitFor(() =>
      expect(screen.getByText(/Não foi possível alcançar o Mac/)).toBeInTheDocument(),
    );
  });

  it("mostra as versões de app, agente e protocolo", () => {
    renderScreen();
    expect(screen.getByText(/App: 0\.1\.0/)).toBeInTheDocument();
    expect(screen.getByText(/Agente: 0\.1\.0/)).toBeInTheDocument();
    expect(screen.getByText(/Protocolo: v1/)).toBeInTheDocument();
  });

  it("chama onClose ao voltar ao painel", () => {
    const props = renderScreen();
    fireEvent.click(screen.getByRole("button", { name: "Voltar ao painel" }));
    expect(props.onClose).toHaveBeenCalledTimes(1);
  });
});
