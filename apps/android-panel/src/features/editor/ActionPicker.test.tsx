import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ActionPicker from "./ActionPicker";
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
    "system.lock": {
      id: "system.lock",
      label: "Bloquear Mac",
      icon: "lock",
      kind: "screen_lock",
      requireLongPress: true,
    },
  };
}

function bigCatalog(): Record<string, ActionSummary> {
  const result = catalog();
  const names = ["Notion", "Spotify", "Figma", "Slack", "Discord", "Postman", "Docker", "Zoom"];
  for (const name of names) {
    const id = `app:${name.toLowerCase()}`;
    result[id] = { id, label: name, icon: "app", kind: "open_app", requireLongPress: false };
  }
  return result;
}

describe("ActionPicker", () => {
  it("lista as ações do catálogo em ordem alfabética", () => {
    render(<ActionPicker actionsCatalog={catalog()} onPick={vi.fn()} onCancel={vi.fn()} />);
    const items = screen.getAllByRole("listitem").map((li) => li.textContent);
    expect(items[0]).toContain("Bloquear Mac");
    expect(items[1]).toContain("Chrome");
  });

  it("chama onPick com o actionId escolhido", async () => {
    const user = userEvent.setup();
    const onPick = vi.fn();
    render(<ActionPicker actionsCatalog={catalog()} onPick={onPick} onCancel={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /Chrome/ }));
    expect(onPick).toHaveBeenCalledWith("app.chrome");
  });

  it("mostra uma mensagem quando não há ações disponíveis", () => {
    render(<ActionPicker actionsCatalog={{}} onPick={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText(/Nenhuma ação disponível/)).toBeInTheDocument();
  });

  it("chama onCancel ao clicar em cancelar", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();
    render(<ActionPicker actionsCatalog={catalog()} onPick={vi.fn()} onCancel={onCancel} />);

    await user.click(screen.getByRole("button", { name: "Cancelar" }));
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it("não mostra busca com poucas ações", () => {
    render(<ActionPicker actionsCatalog={catalog()} onPick={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.queryByLabelText("Buscar app ou ação")).not.toBeInTheDocument();
  });

  it("filtra a lista pelo texto buscado quando há muitas ações", async () => {
    const user = userEvent.setup();
    render(<ActionPicker actionsCatalog={bigCatalog()} onPick={vi.fn()} onCancel={vi.fn()} />);

    await user.type(screen.getByLabelText("Buscar app ou ação"), "notion");

    expect(screen.getByRole("button", { name: /Notion/ })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Chrome/ })).not.toBeInTheDocument();
  });

  it("mostra uma mensagem quando a busca não encontra nada", async () => {
    const user = userEvent.setup();
    render(<ActionPicker actionsCatalog={bigCatalog()} onPick={vi.fn()} onCancel={vi.fn()} />);

    await user.type(screen.getByLabelText("Buscar app ou ação"), "app-inexistente");

    expect(screen.getByText(/Nada encontrado/)).toBeInTheDocument();
  });

  it("renderiza o ícone real quando a ação/app tem iconUrl", () => {
    const withIcon: Record<string, ActionSummary> = {
      ...catalog(),
      "app:notion": {
        id: "app:notion",
        label: "Notion",
        icon: "app",
        kind: "open_app",
        requireLongPress: false,
        iconUrl: "blob:fake-notion-icon",
      },
    };
    render(<ActionPicker actionsCatalog={withIcon} onPick={vi.fn()} onCancel={vi.fn()} />);
    const img = screen.getByRole("button", { name: /Notion/ }).querySelector("img");
    expect(img).toHaveAttribute("src", "blob:fake-notion-icon");
  });
});
