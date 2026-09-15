import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import EditorScreen from "./EditorScreen";
import type { DashboardConfig } from "../../storage/layout";
import type { ActionSummary } from "../../protocol/httpClient";

function catalog(): Record<string, ActionSummary> {
  return {
    "app.chrome": {
      id: "app.chrome",
      label: "Chrome",
      icon: "chrome",
      kind: "open_app",
      requireLongPress: false,
    },
    "app.whatsapp": {
      id: "app.whatsapp",
      label: "WhatsApp",
      icon: "message-circle",
      kind: "open_app",
      requireLongPress: false,
    },
    "system.lock": {
      id: "system.lock",
      label: "Bloquear Mac",
      icon: "lock",
      kind: "screen_lock",
      requireLongPress: true,
    },
  };
}

function makeLayout(): DashboardConfig {
  return {
    schemaVersion: 1,
    activeProfileId: "default",
    profiles: [
      {
        id: "default",
        name: "Padrão",
        pages: [
          {
            id: "page-1",
            name: "Página 1",
            columns: 4,
            rows: 2,
            buttons: [
              { id: "b1", actionId: "app.chrome", position: 0, requireLongPress: false },
              { id: "b2", actionId: "system.lock", position: 1, requireLongPress: true },
            ],
          },
          {
            id: "page-2",
            name: "Página 2",
            columns: 4,
            rows: 2,
            buttons: [
              { id: "b3", actionId: "app.whatsapp", position: 0, requireLongPress: false },
              { id: "b4", actionId: "app.whatsapp", position: 1, requireLongPress: false },
              { id: "b5", actionId: "app.whatsapp", position: 2, requireLongPress: false },
              { id: "b6", actionId: "app.whatsapp", position: 3, requireLongPress: false },
              { id: "b7", actionId: "app.whatsapp", position: 4, requireLongPress: false },
              { id: "b8", actionId: "app.whatsapp", position: 5, requireLongPress: false },
              { id: "b9", actionId: "app.whatsapp", position: 6, requireLongPress: false },
              { id: "b10", actionId: "app.whatsapp", position: 7, requireLongPress: false },
            ],
          },
        ],
      },
    ],
  };
}

function grid() {
  return document.querySelector(".dp-editor__grid") as HTMLElement;
}

describe("EditorScreen", () => {
  it("adiciona um botão em um slot vazio escolhendo uma ação do catálogo", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(async (_next: DashboardConfig) => {});
    render(
      <EditorScreen
        layout={makeLayout()}
        actionsCatalog={catalog()}
        onSave={onSave}
        onRestoreDefault={vi.fn()}
        onClose={vi.fn()}
      />,
    );

    const slots = within(grid()).getAllByRole("button");
    // slot 2 (índice 2) está vazio na Página 1
    await user.click(slots[2]);
    await user.click(screen.getByRole("button", { name: /WhatsApp/ }));

    expect(onSave).toHaveBeenCalledTimes(1);
    const saved = onSave.mock.calls[0][0] as DashboardConfig;
    const newButton = saved.profiles[0].pages[0].buttons.find((b) => b.position === 2);
    expect(newButton?.actionId).toBe("app.whatsapp");
  });

  it("edita o nome de um botão existente", async () => {
    // Harness com estado próprio: EditorScreen é controlado (não guarda
    // cópia do layout), então simular a digitação de verdade exige que o
    // `layout` volte atualizado a cada onSave, como o App real faz via o
    // hook useDeskPanelConnection.
    function Harness() {
      const [layout, setLayout] = useState(makeLayout());
      return (
        <EditorScreen
          layout={layout}
          actionsCatalog={catalog()}
          onSave={async (next) => setLayout(next)}
          onRestoreDefault={vi.fn()}
          onClose={vi.fn()}
        />
      );
    }

    const user = userEvent.setup();
    render(<Harness />);

    const slots = within(grid()).getAllByRole("button");
    await user.click(slots[0]); // botão Chrome

    const nameInput = screen.getByLabelText("Nome no botão");
    await user.type(nameInput, "Navegador");

    expect(nameInput).toHaveValue("Navegador");
  });

  it("remove um botão", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(async (_next: DashboardConfig) => {});
    render(
      <EditorScreen
        layout={makeLayout()}
        actionsCatalog={catalog()}
        onSave={onSave}
        onRestoreDefault={vi.fn()}
        onClose={vi.fn()}
      />,
    );

    const slots = within(grid()).getAllByRole("button");
    await user.click(slots[0]);
    await user.click(screen.getByRole("button", { name: "Remover botão" }));

    const saved = onSave.mock.calls.at(-1)?.[0] as DashboardConfig;
    expect(saved.profiles[0].pages[0].buttons.some((b) => b.id === "b1")).toBe(false);
  });

  it("não permite destravar o toque prolongado de uma ação perigosa", async () => {
    const user = userEvent.setup();
    render(
      <EditorScreen
        layout={makeLayout()}
        actionsCatalog={catalog()}
        onSave={vi.fn(async () => {})}
        onRestoreDefault={vi.fn()}
        onClose={vi.fn()}
      />,
    );

    const slots = within(grid()).getAllByRole("button");
    await user.click(slots[1]); // botão "Bloquear Mac" (system.lock)

    const checkbox = screen.getByLabelText("Exigir toque prolongado") as HTMLInputElement;
    expect(checkbox.checked).toBe(true);
    expect(checkbox.disabled).toBe(true);
  });

  it("reorganiza botões trocando dois slots no modo reorganizar", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(async (_next: DashboardConfig) => {});
    render(
      <EditorScreen
        layout={makeLayout()}
        actionsCatalog={catalog()}
        onSave={onSave}
        onRestoreDefault={vi.fn()}
        onClose={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Reorganizar botões" }));
    const slots = within(grid()).getAllByRole("button");
    await user.click(slots[0]); // seleciona Chrome (posição 0)
    await user.click(slots[1]); // troca com Bloquear Mac (posição 1)

    const saved = onSave.mock.calls.at(-1)?.[0] as DashboardConfig;
    const page1 = saved.profiles[0].pages[0];
    expect(page1.buttons.find((b) => b.id === "b1")?.position).toBe(1);
    expect(page1.buttons.find((b) => b.id === "b2")?.position).toBe(0);
  });

  it("cria uma nova página e permite excluir uma página existente", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(async (_next: DashboardConfig) => {});
    render(
      <EditorScreen
        layout={makeLayout()}
        actionsCatalog={catalog()}
        onSave={onSave}
        onRestoreDefault={vi.fn()}
        onClose={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "+ Nova página" }));
    let saved = onSave.mock.calls.at(-1)?.[0] as DashboardConfig;
    expect(saved.profiles[0].pages).toHaveLength(3);

    await user.click(screen.getByRole("button", { name: "Excluir página Página 2" }));
    await user.click(screen.getByRole("button", { name: "Excluir" }));

    saved = onSave.mock.calls.at(-1)?.[0] as DashboardConfig;
    expect(saved.profiles[0].pages.some((p) => p.id === "page-2")).toBe(false);
  });

  it("não permite excluir a última página restante", () => {
    const layout = makeLayout();
    layout.profiles[0].pages = [layout.profiles[0].pages[0]];

    render(
      <EditorScreen
        layout={layout}
        actionsCatalog={catalog()}
        onSave={vi.fn(async () => {})}
        onRestoreDefault={vi.fn()}
        onClose={vi.fn()}
      />,
    );

    expect(screen.queryByLabelText(/Excluir página/)).not.toBeInTheDocument();
  });

  it("mostra erro ao tentar mover um botão para uma página cheia", async () => {
    const user = userEvent.setup();
    render(
      <EditorScreen
        layout={makeLayout()}
        actionsCatalog={catalog()}
        onSave={vi.fn(async () => {})}
        onRestoreDefault={vi.fn()}
        onClose={vi.fn()}
      />,
    );

    const slots = within(grid()).getAllByRole("button");
    await user.click(slots[0]); // Chrome, na Página 1

    await user.selectOptions(screen.getByLabelText("Mover para página"), "page-2");

    expect(await screen.findByRole("alert")).toHaveTextContent("A página de destino está cheia");
  });

  it("pede confirmação antes de restaurar o layout padrão", async () => {
    const user = userEvent.setup();
    const onRestoreDefault = vi.fn(async () => {});
    render(
      <EditorScreen
        layout={makeLayout()}
        actionsCatalog={catalog()}
        onSave={vi.fn(async () => {})}
        onRestoreDefault={onRestoreDefault}
        onClose={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Restaurar layout padrão" }));
    expect(onRestoreDefault).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "Restaurar" }));
    expect(onRestoreDefault).toHaveBeenCalledTimes(1);
  });

  it("chama onClose ao clicar em 'Voltar ao painel'", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    render(
      <EditorScreen
        layout={makeLayout()}
        actionsCatalog={catalog()}
        onSave={vi.fn(async () => {})}
        onRestoreDefault={vi.fn()}
        onClose={onClose}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Voltar ao painel" }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
