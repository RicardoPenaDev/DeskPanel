import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import PairingScreen from "./PairingScreen";

const initialConfig = {
  deviceId: "device-1",
  deviceName: "Moto G60 do Ricardo",
  host: "",
  port: 38121,
};

describe("PairingScreen", () => {
  it("desabilita 'Parear' até IP, porta e código de 6 dígitos serem válidos", async () => {
    const user = userEvent.setup();
    render(
      <PairingScreen
        initialConfig={initialConfig}
        testConnection={vi.fn()}
        pair={vi.fn()}
        connectionError={null}
      />,
    );

    const pairButton = screen.getByRole("button", { name: "Parear" });
    expect(pairButton).toBeDisabled();

    await user.type(screen.getByLabelText("IP ou hostname do Mac"), "192.168.1.50");
    await user.type(screen.getByLabelText("Código de pareamento"), "483921");

    expect(pairButton).toBeEnabled();
  });

  it("filtra caracteres não numéricos e limita o código a 6 dígitos", async () => {
    const user = userEvent.setup();
    render(
      <PairingScreen
        initialConfig={initialConfig}
        testConnection={vi.fn()}
        pair={vi.fn()}
        connectionError={null}
      />,
    );

    const codeInput = screen.getByLabelText("Código de pareamento");
    await user.type(codeInput, "4a8-3.9²21999");

    expect(codeInput).toHaveValue("483921");
  });

  it("mostra o resultado de 'Testar conexão'", async () => {
    const user = userEvent.setup();
    const testConnection = vi.fn(async () => ({
      ok: true as const,
      data: { status: "ok", service: "deskpanel-agent", protocolVersion: 1 },
    }));

    render(
      <PairingScreen
        initialConfig={initialConfig}
        testConnection={testConnection}
        pair={vi.fn()}
        connectionError={null}
      />,
    );

    await user.type(screen.getByLabelText("IP ou hostname do Mac"), "192.168.1.50");
    await user.click(screen.getByRole("button", { name: "Testar conexão" }));

    expect(testConnection).toHaveBeenCalledWith("192.168.1.50", 38121);
    await waitFor(() => expect(screen.getByText(/Conectado: deskpanel-agent/)).toBeInTheDocument());
  });

  it("mostra a mensagem de erro quando pair() falha", async () => {
    const user = userEvent.setup();
    const pair = vi.fn(async () => ({ ok: false as const, message: "Código incorreto." }));

    render(
      <PairingScreen
        initialConfig={initialConfig}
        testConnection={vi.fn()}
        pair={pair}
        connectionError={null}
      />,
    );

    await user.type(screen.getByLabelText("IP ou hostname do Mac"), "192.168.1.50");
    await user.type(screen.getByLabelText("Código de pareamento"), "000000");
    await user.click(screen.getByRole("button", { name: "Parear" }));

    expect(pair).toHaveBeenCalledWith({
      deviceName: "Moto G60 do Ricardo",
      host: "192.168.1.50",
      port: 38121,
      code: "000000",
    });
    await waitFor(() => expect(screen.getByRole("alert")).toHaveTextContent("Código incorreto."));
  });

  it("nunca deixa o token de acesso aparecer na tela", () => {
    render(
      <PairingScreen
        initialConfig={initialConfig}
        testConnection={vi.fn()}
        pair={vi.fn()}
        connectionError="dispositivo revogado"
      />,
    );

    expect(document.body.textContent).not.toMatch(/accessToken|token-/i);
  });
});
